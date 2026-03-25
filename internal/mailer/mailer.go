package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"gopkg.in/gomail.v2"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	BaseURL  string // e.g. http://localhost:8080
}

type Mailer struct {
	cfg Config
	db  *db.DB
}

func New(cfg Config, database *db.DB) *Mailer {
	return &Mailer{cfg: cfg, db: database}
}

// emailData is passed into the HTML template
type emailData struct {
	FirstName     string
	TrackingPixel string
	ClickURL      string
	SenderName    string
}

var emailTemplate = template.Must(template.New("email").Parse(`<!DOCTYPE html>
<html>
<body style="font-family:Segoe UI,sans-serif;color:#333;max-width:600px;margin:auto;padding:32px;">
  <p>Hi {{.FirstName}},</p>

  <p>We detected unusual sign-in activity on your account.
  Please verify your identity immediately to avoid suspension.</p>

  <p style="text-align:center;margin:32px 0;">
    <a href="{{.ClickURL}}"
       style="background:#0067b8;color:white;padding:12px 28px;text-decoration:none;border-radius:4px;font-size:15px;">
      Verify My Account
    </a>
  </p>

  <p style="color:#888;font-size:12px;">
    If you did not request this, ignore this email.<br/>
    — {{.SenderName}}
  </p>

  <!-- tracking pixel -->
  <img src="{{.TrackingPixel}}" width="1" height="1" style="display:none;" />
</body>
</html>`))

// SendCampaign sends emails to all targets in a campaign
func (m *Mailer) SendCampaign(ctx context.Context, campaignID string) error {
	campaign, err := m.db.GetCampaign(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("get campaign: %w", err)
	}

	targets, err := m.db.GetTargetsByCampaign(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("get targets: %w", err)
	}

	dialer := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.Username, m.cfg.Password)

	for _, target := range targets {
		body, err := m.buildBody(target, campaign.SenderName)
		if err != nil {
			log.Printf("build email for %s: %v", target.Email, err)
			continue
		}

		msg := gomail.NewMessage()
		msg.SetHeader("From", fmt.Sprintf("%s <%s>", campaign.SenderName, campaign.SenderEmail))
		msg.SetHeader("To", target.Email)
		msg.SetHeader("Subject", campaign.Subject)
		msg.SetBody("text/html", body)

		var sendErr error
		for attempt := 1; attempt <= 3; attempt++ {
			sendErr = dialer.DialAndSend(msg)
			if sendErr == nil {
				break
			}
			log.Printf("attempt %d failed for %s: %v — retrying in %ds", attempt, target.Email, sendErr, attempt*10)
			time.Sleep(time.Duration(attempt*10) * time.Second)
		}
		if sendErr != nil {
			log.Printf("✗ gave up on %s after 3 attempts", target.Email)
			continue
		}

		// log the send event
		m.db.LogEvent(ctx, db.Event{
			CampaignID: target.CampaignID,
			TargetID:   target.ID,
			EventType:  "email_sent",
		})

		log.Printf("✓ sent to %s", target.Email)
	}

	return nil
}

func (m *Mailer) buildBody(target db.Target, senderName string) (string, error) {
	data := emailData{
		FirstName:     target.FirstName,
		TrackingPixel: fmt.Sprintf("%s/t/open/%s.png", m.cfg.BaseURL, target.Token),
		ClickURL:      fmt.Sprintf("%s/t/click/%s", m.cfg.BaseURL, target.Token),
		SenderName:    senderName,
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
