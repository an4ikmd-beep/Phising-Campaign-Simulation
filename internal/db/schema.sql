-- internal/db/schema.sql

CREATE TABLE IF NOT EXISTS campaigns (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    subject      TEXT NOT NULL,
    sender_name  TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'draft',
    created_at   DATETIME NOT NULL,
    launched_at  DATETIME
);

CREATE TABLE IF NOT EXISTS targets (
    id           TEXT PRIMARY KEY,
    campaign_id  TEXT NOT NULL REFERENCES campaigns(id),
    email        TEXT NOT NULL,
    first_name   TEXT,
    last_name    TEXT,
    token        TEXT NOT NULL UNIQUE,  -- used in tracking URLs
    sent_at      DATETIME
);

CREATE TABLE IF NOT EXISTS events (
    id           TEXT PRIMARY KEY,
    campaign_id  TEXT NOT NULL REFERENCES campaigns(id),
    target_id    TEXT NOT NULL REFERENCES targets(id),
    event_type   TEXT NOT NULL,  -- email_sent | opened | clicked | submitted
    ip           TEXT,
    user_agent   TEXT,
    timestamp    DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_campaign ON events(campaign_id);
CREATE INDEX IF NOT EXISTS idx_events_target   ON events(target_id);
CREATE INDEX IF NOT EXISTS idx_targets_token   ON targets(token);