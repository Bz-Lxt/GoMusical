-- GoMusical schema v1. Serialized by pg_advisory_lock in migrate.go.

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('CREATOR', 'LISTENER', 'ADMIN')),
    avatar_url    TEXT NOT NULL DEFAULT '',
    bio           TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    csrf_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

CREATE TABLE IF NOT EXISTS albums (
    id          TEXT PRIMARY KEY,
    creator_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    cover_key   TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS asset_blobs (
    sha256      TEXT PRIMARY KEY,
    storage_key TEXT NOT NULL,
    size_bytes  BIGINT NOT NULL,
    mime        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS tracks (
    id               TEXT PRIMARY KEY,
    creator_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id         TEXT REFERENCES albums(id) ON DELETE SET NULL,
    title            TEXT NOT NULL,
    display_filename TEXT NOT NULL,
    duration_ms      INT NOT NULL DEFAULT 0,
    format           TEXT NOT NULL DEFAULT '',
    content_sha256   TEXT NOT NULL DEFAULT '',
    storage_key      TEXT NOT NULL DEFAULT '',
    size_bytes       BIGINT NOT NULL DEFAULT 0,
    preview_seconds  INT NOT NULL DEFAULT 30,
    paid_download    BOOLEAN NOT NULL DEFAULT TRUE,
    paid_price_cents INT NOT NULL DEFAULT 900,
    fan_only         BOOLEAN NOT NULL DEFAULT FALSE,
    fan_download     BOOLEAN NOT NULL DEFAULT FALSE,
    play_count       BIGINT NOT NULL DEFAULT 0,
    sponsor_cents    BIGINT NOT NULL DEFAULT 0,
    transcode_status TEXT NOT NULL DEFAULT 'pending',
    transcode_error  TEXT NOT NULL DEFAULT '',
    peaks_key        TEXT NOT NULL DEFAULT '',
    hls_dir          TEXT NOT NULL DEFAULT '',
    cover_key        TEXT NOT NULL DEFAULT '',
    segment_count    INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tracks_creator ON tracks(creator_id);
CREATE INDEX IF NOT EXISTS idx_tracks_sha ON tracks(content_sha256);

CREATE TABLE IF NOT EXISTS comments (
    id            TEXT PRIMARY KEY,
    track_id      TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    timestamp_ms  INT NOT NULL,
    body          TEXT NOT NULL,
    likes         INT NOT NULL DEFAULT 0,
    pinned        BOOLEAN NOT NULL DEFAULT FALSE,
    hidden        BOOLEAN NOT NULL DEFAULT FALSE,
    reply         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_comments_track ON comments(track_id, timestamp_ms);

CREATE TABLE IF NOT EXISTS orders (
    id           TEXT PRIMARY KEY,
    order_no     TEXT NOT NULL UNIQUE,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id     TEXT,
    creator_id   TEXT,
    kind         TEXT NOT NULL CHECK (kind IN ('TRACK_SPONSOR', 'FAN_SUB')),
    amount_cents INT NOT NULL,
    status       TEXT NOT NULL,
    provider     TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    paid_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS grants (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   TEXT REFERENCES tracks(id) ON DELETE CASCADE,
    creator_id TEXT,
    kind       TEXT NOT NULL CHECK (kind IN ('PAID_DOWNLOAD', 'FAN_ONLY', 'FAN_DOWNLOAD')),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_grants_user_track ON grants(user_id, track_id);

CREATE TABLE IF NOT EXISTS subscriptions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, creator_id)
);

CREATE TABLE IF NOT EXISTS download_tickets (
    nonce        TEXT PRIMARY KEY,
    grant_id     TEXT NOT NULL,
    user_id      TEXT NOT NULL,
    track_id     TEXT NOT NULL,
    max_uses     INT NOT NULL,
    uses         INT NOT NULL DEFAULT 0,
    bytes_done   BIGINT NOT NULL DEFAULT 0,
    revoked      BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transcode_jobs (
    id         TEXT PRIMARY KEY,
    track_id   TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    status     TEXT NOT NULL,
    progress   INT NOT NULL DEFAULT 0,
    error      TEXT NOT NULL DEFAULT '',
    attempts   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id         TEXT PRIMARY KEY,
    actor_id   TEXT NOT NULL DEFAULT '',
    action     TEXT NOT NULL,
    reason     TEXT NOT NULL,
    meta       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action, created_at DESC);

CREATE TABLE IF NOT EXISTS upload_sessions (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    filename     TEXT NOT NULL,
    sha256       TEXT NOT NULL,
    size_bytes   BIGINT NOT NULL,
    chunk_size   INT NOT NULL,
    received     BOOLEAN[] NOT NULL,
    tmp_key      TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL
);
