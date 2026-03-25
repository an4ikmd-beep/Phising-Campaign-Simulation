package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Campaign struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Subject     string     `db:"subject"`
	SenderName  string     `db:"sender_name"`
	SenderEmail string     `db:"sender_email"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	LaunchedAt  *time.Time `db:"launched_at"`
}

type Target struct {
	ID         string     `db:"id"`
	CampaignID string     `db:"campaign_id"`
	Email      string     `db:"email"`
	FirstName  string     `db:"first_name"`
	LastName   string     `db:"last_name"`
	Token      string     `db:"token"`
	SentAt     *time.Time `db:"sent_at"`
}

type Event struct {
	ID         string    `db:"id"`
	CampaignID string    `db:"campaign_id"`
	TargetID   string    `db:"target_id"`
	EventType  string    `db:"event_type"`
	IP         string    `db:"ip"`
	UserAgent  string    `db:"user_agent"`
	Timestamp  time.Time `db:"timestamp"`
}

func (db *DB) CreateCampaign(ctx context.Context, c Campaign) error {
	c.ID = uuid.NewString()
	c.CreatedAt = time.Now()
	_, err := db.conn.NamedExecContext(ctx, `
        INSERT INTO campaigns (id, name, subject, sender_name, sender_email, status, created_at)
        VALUES (:id, :name, :subject, :sender_name, :sender_email, :status, :created_at)
    `, c)
	return err
}

func (db *DB) GetCampaign(ctx context.Context, id string) (*Campaign, error) {
	var c Campaign
	err := db.conn.GetContext(ctx, &c, `SELECT * FROM campaigns WHERE id = ?`, id)
	return &c, err
}

func (db *DB) ListCampaigns(ctx context.Context) ([]Campaign, error) {
	var list []Campaign
	err := db.conn.SelectContext(ctx, &list, `SELECT * FROM campaigns ORDER BY created_at DESC`)
	return list, err
}

func (db *DB) CreateTarget(ctx context.Context, t Target) error {
	t.ID = uuid.NewString()
	t.Token = uuid.NewString() // unique tracking token
	_, err := db.conn.NamedExecContext(ctx, `
        INSERT INTO targets (id, campaign_id, email, first_name, last_name, token)
        VALUES (:id, :campaign_id, :email, :first_name, :last_name, :token)
    `, t)
	return err
}

// GetTargetByToken is used by the tracker to look up who clicked
func (db *DB) GetTargetByToken(ctx context.Context, token string) (*Target, error) {
	var t Target
	err := db.conn.GetContext(ctx, &t, `SELECT * FROM targets WHERE token = ?`, token)
	return &t, err
}

func (db *DB) GetTargetsByCampaign(ctx context.Context, campaignID string) ([]Target, error) {
	var list []Target
	err := db.conn.SelectContext(ctx, &list, `SELECT * FROM targets WHERE campaign_id = ?`, campaignID)
	return list, err
}

func (db *DB) LogEvent(ctx context.Context, e Event) error {
	e.ID = uuid.NewString()
	e.Timestamp = time.Now()
	_, err := db.conn.NamedExecContext(ctx, `
        INSERT INTO events (id, campaign_id, target_id, event_type, ip, user_agent, timestamp)
        VALUES (:id, :campaign_id, :target_id, :event_type, :ip, :user_agent, :timestamp)
    `, e)
	return err
}

func (db *DB) GetEventsByCampaign(ctx context.Context, campaignID string) ([]Event, error) {
	var list []Event
	err := db.conn.SelectContext(ctx, &list, `
        SELECT * FROM events WHERE campaign_id = ? ORDER BY timestamp ASC
    `, campaignID)
	return list, err
}

// returns a map of event_type → count for a campaign
func (db *DB) CountByType(ctx context.Context, campaignID string) (map[string]int, error) {
	rows, err := db.conn.QueryContext(ctx, `
        SELECT event_type, COUNT(*) FROM events
        WHERE campaign_id = ?
        GROUP BY event_type
    `, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var eventType string
		var count int
		rows.Scan(&eventType, &count)
		counts[eventType] = count
	}
	return counts, nil
}
