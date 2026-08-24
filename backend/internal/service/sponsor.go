package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gomusical/internal/clock"
	"gomusical/internal/config"
	"gomusical/internal/httpx"
	"gomusical/internal/hmacx"
	"gomusical/internal/model"
	"gomusical/internal/payment"
	"gomusical/internal/repo"
)

type Sponsor struct {
	Repos    *repo.Repos
	Pay      payment.Provider
	Cfg      config.Config
	Access   *Access
}

type CheckoutOut struct {
	Order       model.Order `json:"order"`
	CheckoutURL string      `json:"checkoutUrl"`
	Paid        bool        `json:"paid"`
}

func (s *Sponsor) CreateTrackSponsor(ctx context.Context, userID, trackID string, amount int) (*CheckoutOut, error) {
	if userID == "" {
		return nil, httpx.ErrUnauthorized
	}
	tr, err := s.Repos.TrackByID(ctx, trackID)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		amount = tr.PaidPriceCents
	}
	if amount < tr.PaidPriceCents {
		return nil, httpx.New(422, "amount_low", "赞助金额低于作品设定")
	}
	o := &model.Order{
		ID: uuid.NewString(), OrderNo: repo.NewOrderNo(), UserID: userID,
		TrackID: &tr.ID, Kind: model.KindTrackSponsor, AmountCents: amount,
		Status: model.OrderPending, Provider: s.Pay.Name(), CreatedAt: clock.Now(),
	}
	if err := s.Repos.CreateOrder(ctx, o); err != nil {
		return nil, err
	}
	res, err := s.Pay.Charge(ctx, payment.ChargeRequest{
		OrderNo: o.OrderNo, AmountCents: amount, UserID: userID, Subject: tr.Title,
		ReturnURL: s.Cfg.PublicOrigin + "/pay/result",
	})
	if err != nil && res.Status != "failed" {
		o.Status = model.OrderFailed
		s.Repos.Audit(ctx, userID, "pay.fail", err.Error(), map[string]any{"order": o.OrderNo})
		return &CheckoutOut{Order: *o, CheckoutURL: res.CheckoutURL}, err
	}
	out := &CheckoutOut{Order: *o, CheckoutURL: res.CheckoutURL}
	if res.Status == "paid" {
		if err := s.Fulfill(ctx, o.OrderNo); err != nil {
			return nil, err
		}
		o.Status = model.OrderPaid
		o.PaidAt = res.PaidAt
		out.Paid = true
		out.Order = *o
	}
	return out, nil
}

func (s *Sponsor) CreateFanSub(ctx context.Context, userID, creatorID string, amount int) (*CheckoutOut, error) {
	if userID == "" {
		return nil, httpx.ErrUnauthorized
	}
	if amount <= 0 {
		amount = 1800
	}
	o := &model.Order{
		ID: uuid.NewString(), OrderNo: repo.NewOrderNo(), UserID: userID,
		CreatorID: &creatorID, Kind: model.KindFanSub, AmountCents: amount,
		Status: model.OrderPending, Provider: s.Pay.Name(), CreatedAt: clock.Now(),
	}
	if err := s.Repos.CreateOrder(ctx, o); err != nil {
		return nil, err
	}
	res, err := s.Pay.Charge(ctx, payment.ChargeRequest{
		OrderNo: o.OrderNo, AmountCents: amount, UserID: userID, Subject: "粉丝月度订阅",
	})
	if err != nil {
		return &CheckoutOut{Order: *o}, err
	}
	out := &CheckoutOut{Order: *o, CheckoutURL: res.CheckoutURL}
	if res.Status == "paid" {
		if err := s.Fulfill(ctx, o.OrderNo); err != nil {
			return nil, err
		}
		o.Status = model.OrderPaid
		out.Paid = true
		out.Order = *o
	}
	return out, nil
}

// Fulfill is idempotent: same orderNo only grants once.
func (s *Sponsor) Fulfill(ctx context.Context, orderNo string) error {
	o, err := s.Repos.OrderByNo(ctx, orderNo)
	if err != nil {
		return err
	}
	first, err := s.Repos.MarkOrderPaid(ctx, orderNo, clock.Now())
	if err != nil {
		return err
	}
	if !first {
		return nil
	}
	if o.Kind == model.KindTrackSponsor && o.TrackID != nil {
		g := &model.Grant{
			ID: uuid.NewString(), UserID: o.UserID, TrackID: o.TrackID,
			Kind: model.TierPaid, CreatedAt: clock.Now(),
		}
		if err := s.Repos.CreateGrant(ctx, g); err != nil {
			return err
		}
		tr, err := s.Repos.TrackByID(ctx, *o.TrackID)
		if err == nil {
			tr.SponsorCents += int64(o.AmountCents)
			tr.UpdatedAt = clock.Now()
			_ = s.Repos.UpdateTrack(ctx, tr)
		}
		s.Repos.Audit(ctx, o.UserID, "grant.issue", "paid_download", map[string]any{"grant": g.ID, "order": orderNo})
	}
	if o.Kind == model.KindFanSub && o.CreatorID != nil {
		exp := clock.Now().Add(30 * 24 * time.Hour)
		sub := &model.Subscription{
			ID: uuid.NewString(), UserID: o.UserID, CreatorID: *o.CreatorID,
			ExpiresAt: exp, CreatedAt: clock.Now(),
		}
		if err := s.Repos.UpsertSub(ctx, sub); err != nil {
			return err
		}
		s.Repos.Audit(ctx, o.UserID, "sub.issue", "fan_month", map[string]any{"creator": *o.CreatorID})
	}
	return nil
}

func (s *Sponsor) Callback(ctx context.Context, payload map[string]string) error {
	orderNo, ok, err := s.Pay.VerifyCallback(ctx, payload)
	if err != nil {
		return err
	}
	if !ok {
		s.Repos.Audit(ctx, "", "pay.callback_fail", "rejected", payload)
		return httpx.New(400, "pay_rejected", "支付回调校验失败")
	}
	return s.Fulfill(ctx, orderNo)
}

func (s *Sponsor) IssueTicket(ctx context.Context, userID, trackID string) (string, hmacx.Ticket, error) {
	tr, err := s.Repos.TrackByID(ctx, trackID)
	if err != nil {
		return "", hmacx.Ticket{}, err
	}
	d := s.Access.Decide(ctx, userID, tr)
	if !d.CanDownload {
		return "", hmacx.Ticket{}, httpx.ErrForbidden
	}
	grantID := d.GrantID
	if grantID == "" {
		g := &model.Grant{
			ID: uuid.NewString(), UserID: userID, TrackID: &tr.ID,
			Kind: model.TierPaid, CreatedAt: clock.Now(),
		}
		if err := s.Repos.CreateGrant(ctx, g); err != nil {
			return "", hmacx.Ticket{}, err
		}
		grantID = g.ID
	}
	ttl := s.Cfg.TicketTTL
	if ttl > s.Cfg.TicketTTLMax {
		ttl = s.Cfg.TicketTTLMax
	}
	tk := hmacx.NewTicket(tr.ID, userID, grantID, "lossless", ttl, s.Cfg.TicketMaxUses)
	raw, err := hmacx.SignTicket(tk, s.Cfg.HMACSecret)
	if err != nil {
		return "", hmacx.Ticket{}, err
	}
	row := &model.DownloadTicketRow{
		Nonce: tk.Nonce, GrantID: grantID, UserID: userID, TrackID: tr.ID,
		MaxUses: tk.MaxUses, ExpiresAt: time.Unix(tk.Exp, 0), CreatedAt: clock.Now(),
	}
	if err := s.Repos.InsertTicket(ctx, row); err != nil {
		return "", hmacx.Ticket{}, err
	}
	s.Repos.Audit(ctx, userID, "ticket.issue", "ok", map[string]any{"nonce": tk.Nonce, "track": tr.ID})
	return raw, tk, nil
}
