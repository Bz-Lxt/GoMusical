package auditx

// Reason codes written to audit_logs.reason. Keep stable for dashboards.
const (
	TicketIssue     = "ticket.issue"
	TicketVerifyFail = "ticket.verify_fail"
	TicketRevoke    = "ticket.revoke"
	TicketExpired   = "ticket.expired"
	TicketTamper    = "ticket.tamper"
	DownloadOK      = "download.complete"
	Download429     = "download.limited"
	StreamDenied    = "stream.denied"
	StreamOpen      = "stream.open"
	PayFail         = "pay.fail"
	PayCallbackFail = "pay.callback_fail"
	GrantIssue      = "grant.issue"
	SubIssue        = "sub.issue"
	CommentOOB      = "comment.window"
	UploadHashFail  = "upload.hash_mismatch"
	TranscodeFail   = "transcode.fail"
	RefererDeny     = "referer.denied"
)

var Catalog = map[string]string{
	TicketIssue:      "签发下载凭证",
	TicketVerifyFail: "凭证验签失败",
	TicketRevoke:     "凭证被吊销",
	TicketExpired:    "凭证过期",
	TicketTamper:     "凭证被篡改",
	DownloadOK:       "无损下载完成（累计≥95%）",
	Download429:      "并发或日限额拦截",
	StreamDenied:     "切片超出授权窗口",
	StreamOpen:       "打开 HLS 会话",
	PayFail:          "支付下单失败",
	PayCallbackFail:  "支付回调拒绝",
	GrantIssue:       "赞助授权落地",
	SubIssue:         "粉丝订阅落地",
	CommentOOB:       "乐评时间越界",
	UploadHashFail:   "分片合并后哈希不一致",
	TranscodeFail:    "FFmpeg 任务失败",
	RefererDeny:      "防盗链拒绝",
}

func Label(code string) string {
	if s, ok := Catalog[code]; ok {
		return s
	}
	return code
}
