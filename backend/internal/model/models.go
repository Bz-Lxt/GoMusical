package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"displayName"`
	Role         string    `json:"role"`
	AvatarURL    string    `json:"avatarUrl"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Session struct {
	ID        string
	UserID    string
	CSRF      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Album struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creatorId"`
	Title       string    `json:"title"`
	CoverKey    string    `json:"coverKey"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	TrackCount  int       `json:"trackCount"`
}

type AssetBlob struct {
	SHA256     string
	StorageKey string
	SizeBytes  int64
	MIME       string
	CreatedAt  time.Time
}

type Track struct {
	ID              string    `json:"id"`
	CreatorID       string    `json:"creatorId"`
	AlbumID         *string   `json:"albumId"`
	Title           string    `json:"title"`
	DisplayFilename string    `json:"displayFilename"`
	DurationMS      int       `json:"durationMs"`
	Format          string    `json:"format"`
	ContentSHA256   string    `json:"contentSha256"`
	StorageKey      string    `json:"-"`
	SizeBytes       int64     `json:"sizeBytes"`
	PreviewSeconds  int       `json:"previewSeconds"`
	PaidDownload    bool      `json:"paidDownload"`
	PaidPriceCents  int       `json:"paidPriceCents"`
	FanOnly         bool      `json:"fanOnly"`
	FanDownload     bool      `json:"fanDownload"`
	PlayCount       int64     `json:"playCount"`
	SponsorCents    int64     `json:"sponsorCents"`
	TranscodeStatus string    `json:"transcodeStatus"`
	TranscodeError  string    `json:"transcodeError"`
	PeaksKey        string    `json:"-"`
	HLSDir          string    `json:"-"`
	CoverKey        string    `json:"coverKey"`
	SegmentCount    int       `json:"segmentCount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatorName     string    `json:"creatorName,omitempty"`
	AccessTier      string    `json:"accessTier,omitempty"`
	AccessUntilMS   int       `json:"accessUntilMs,omitempty"`
}

type Comment struct {
	ID           string    `json:"id"`
	TrackID      string    `json:"trackId"`
	UserID       string    `json:"userId"`
	AuthorName   string    `json:"authorName"`
	TimestampMS  int       `json:"timestampMs"`
	Body         string    `json:"body"`
	Likes        int       `json:"likes"`
	Pinned       bool      `json:"pinned"`
	Hidden       bool      `json:"hidden"`
	Reply        string    `json:"reply"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Order struct {
	ID          string     `json:"id"`
	OrderNo     string     `json:"orderNo"`
	UserID      string     `json:"userId"`
	TrackID     *string    `json:"trackId"`
	CreatorID   *string    `json:"creatorId"`
	Kind        string     `json:"kind"`
	AmountCents int        `json:"amountCents"`
	Status      string     `json:"status"`
	Provider    string     `json:"provider"`
	CreatedAt   time.Time  `json:"createdAt"`
	PaidAt      *time.Time `json:"paidAt"`
}

type Grant struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	TrackID   *string    `json:"trackId"`
	CreatorID *string    `json:"creatorId"`
	Kind      string     `json:"kind"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

type Subscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	CreatorID string    `json:"creatorId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type DownloadTicketRow struct {
	Nonce      string
	GrantID    string
	UserID     string
	TrackID    string
	MaxUses    int
	Uses       int
	BytesDone  int64
	Revoked    bool
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

type TranscodeJob struct {
	ID        string    `json:"id"`
	TrackID   string    `json:"trackId"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Error     string    `json:"error"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actorId"`
	Action    string    `json:"action"`
	Reason    string    `json:"reason"`
	Meta      string    `json:"meta"`
	CreatedAt time.Time `json:"createdAt"`
}

type UploadSession struct {
	ID         string
	UserID     string
	Filename   string
	SHA256     string
	SizeBytes  int64
	ChunkSize  int
	Received   []bool
	TmpKey     string
	CreatedAt  time.Time
}

const (
	RoleCreator  = "CREATOR"
	RoleListener = "LISTENER"
	RoleAdmin    = "ADMIN"

	TierPreview = "PREVIEW"
	TierPaid    = "PAID_DOWNLOAD"
	TierFan     = "FAN_ONLY"

	JobPending = "pending"
	JobRunning = "running"
	JobReady   = "ready"
	JobFailed  = "failed"

	OrderPending = "pending"
	OrderPaid    = "paid"
	OrderFailed  = "failed"

	KindTrackSponsor = "TRACK_SPONSOR"
	KindFanSub       = "FAN_SUB"
)
