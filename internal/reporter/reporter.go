package reporter

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
)

type Reporter struct {
	DB *db.DB
}

func New(database *db.DB) *Reporter {
	return &Reporter{DB: database}
}

type TargetRow struct {
	FirstName string
	LastName  string
	Email     string
	EmailSent bool
	Opened    bool
	Clicked   bool
	Submitted bool
	RiskLevel string // Low / Medium / High / Critical
}

type TimelineEntry struct {
	Timestamp time.Time
	EventType string
	FirstName string
	LastName  string
	Email     string
	IP        string
}

type Report struct {
	CampaignName   string
	GeneratedAt    time.Time
	TotalTargets   int
	EmailsSent     int
	Opened         int
	Clicked        int
	Submitted      int
	OpenRate       float64
	ClickRate      float64
	SubmissionRate float64
	RiskScore      string
	RiskColor      string
	Targets        []TargetRow
	Timeline       []TimelineEntry
}

func (r *Reporter) Build(ctx context.Context, campaignID string) (*Report, error) {
	campaign, err := r.DB.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}

	targets, err := r.DB.GetTargetsByCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("get targets: %w", err)
	}

	events, err := r.DB.GetEventsByCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}

	// index events by target ID
	type eventSet struct {
		sent, opened, clicked, submitted bool
	}
	byTarget := map[string]*eventSet{}
	for _, t := range targets {
		byTarget[t.ID] = &eventSet{}
	}
	for _, e := range events {
		es := byTarget[e.TargetID]
		if es == nil {
			continue
		}
		switch e.EventType {
		case "email_sent":
			es.sent = true
		case "opened":
			es.opened = true
		case "clicked":
			es.clicked = true
		case "submitted":
			es.submitted = true
		}
	}

	// build target rows
	var rows []TargetRow
	sent, opened, clicked, submitted := 0, 0, 0, 0
	for _, t := range targets {
		es := byTarget[t.ID]
		if es.sent {
			sent++
		}
		if es.opened {
			opened++
		}
		if es.clicked {
			clicked++
		}
		if es.submitted {
			submitted++
		}
		rows = append(rows, TargetRow{
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Email:     t.Email,
			EmailSent: es.sent,
			Opened:    es.opened,
			Clicked:   es.clicked,
			Submitted: es.submitted,
			RiskLevel: targetRisk(es.clicked, es.submitted),
		})
	}

	// build timeline — need target lookup by ID
	targetByID := map[string]db.Target{}
	for _, t := range targets {
		targetByID[t.ID] = t
	}
	var timeline []TimelineEntry
	for _, e := range events {
		t := targetByID[e.TargetID]
		timeline = append(timeline, TimelineEntry{
			Timestamp: e.Timestamp,
			EventType: e.EventType,
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Email:     t.Email,
			IP:        e.IP,
		})
	}

	// rates
	openRate, clickRate, submitRate := 0.0, 0.0, 0.0
	if sent > 0 {
		openRate = float64(opened) / float64(sent) * 100
		clickRate = float64(clicked) / float64(sent) * 100
		submitRate = float64(submitted) / float64(sent) * 100
	}

	risk, color := campaignRisk(clickRate, submitRate)

	return &Report{
		CampaignName:   campaign.Name,
		GeneratedAt:    time.Now(),
		TotalTargets:   len(targets),
		EmailsSent:     sent,
		Opened:         opened,
		Clicked:        clicked,
		Submitted:      submitted,
		OpenRate:       openRate,
		ClickRate:      clickRate,
		SubmissionRate: submitRate,
		RiskScore:      risk,
		RiskColor:      color,
		Targets:        rows,
		Timeline:       timeline,
	}, nil
}

func targetRisk(clicked, submitted bool) string {
	if submitted {
		return "Critical"
	}
	if clicked {
		return "High"
	}
	return "Low"
}

func campaignRisk(clickRate, submitRate float64) (string, string) {
	switch {
	case submitRate > 20 || clickRate > 50:
		return "Critical", "#c0392b"
	case submitRate > 10 || clickRate > 30:
		return "High", "#e67e22"
	case clickRate > 5:
		return "Medium", "#f1c40f"
	default:
		return "Low", "#27ae60"
	}
}

var reportTmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"formatRate": func(f float64) string {
		return fmt.Sprintf("%.1f%%", f)
	},
	"eventLabel": func(s string) string {
		switch s {
		case "email_sent":
			return "📧 Email Sent"
		case "opened":
			return "👁 Opened"
		case "clicked":
			return "🖱 Clicked"
		case "submitted":
			return "⚠️ Submitted Credentials"
		default:
			return s
		}
	},
	"riskBadge": func(s string) template.HTML {
		colors := map[string]string{
			"Critical": "#c0392b",
			"High":     "#e67e22",
			"Medium":   "#f1c40f",
			"Low":      "#27ae60",
		}
		c := colors[s]
		return template.HTML(fmt.Sprintf(
			`<span style="background:%s;color:white;padding:2px 8px;border-radius:4px;font-size:12px;">%s</span>`, c, s,
		))
	},
	"checkmark": func(b bool) string {
		if b {
			return "✓"
		}
		return "—"
	},
}).Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8"/>
<title>Phishing Simulation Report — {{.CampaignName}}</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: Segoe UI, sans-serif; background: #f5f6fa; color: #2d3436; padding: 40px; }
  h1 { font-size: 26px; margin-bottom: 4px; }
  .meta { color: #636e72; font-size: 13px; margin-bottom: 32px; }
  .cards { display: flex; gap: 16px; margin-bottom: 32px; flex-wrap: wrap; }
  .card { background: white; border-radius: 8px; padding: 20px 28px; flex: 1; min-width: 140px;
          box-shadow: 0 1px 4px rgba(0,0,0,.08); }
  .card .val { font-size: 32px; font-weight: 700; margin-bottom: 4px; }
  .card .lbl { font-size: 13px; color: #636e72; }
  .risk-badge { display:inline-block; padding: 6px 18px; border-radius: 6px;
                color: white; font-weight: 600; font-size: 18px; margin-bottom: 32px; }
  h2 { font-size: 18px; margin-bottom: 12px; margin-top: 32px; }
  table { width: 100%; border-collapse: collapse; background: white;
          border-radius: 8px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,.08); }
  th { background: #2d3436; color: white; padding: 10px 14px; text-align: left; font-size: 13px; }
  td { padding: 10px 14px; font-size: 13px; border-bottom: 1px solid #f0f0f0; }
  tr:last-child td { border-bottom: none; }
  tr:hover td { background: #fafafa; }
  .check-yes { color: #27ae60; font-weight: 700; }
  .check-no  { color: #b2bec3; }
  .timeline-entry { background: white; border-radius: 6px; padding: 10px 16px;
                    margin-bottom: 8px; box-shadow: 0 1px 3px rgba(0,0,0,.06);
                    display: flex; gap: 16px; align-items: center; font-size: 13px; }
  .timeline-time { color: #636e72; min-width: 160px; }
  .timeline-event { font-weight: 600; min-width: 200px; }
  .timeline-who { color: #2d3436; }
  .timeline-ip { color: #b2bec3; font-size: 11px; }
</style>
</head>
<body>

<h1>🎣 Phishing Simulation Report</h1>
<div class="meta">Campaign: <strong>{{.CampaignName}}</strong> &nbsp;|&nbsp; Generated: {{formatTime .GeneratedAt}}</div>

<!-- Risk Score -->
<div class="risk-badge" style="background:{{.RiskColor}}">
  Risk Score: {{.RiskScore}}
</div>

<!-- Stat Cards -->
<div class="cards">
  <div class="card"><div class="val">{{.TotalTargets}}</div><div class="lbl">Targets</div></div>
  <div class="card"><div class="val">{{.EmailsSent}}</div><div class="lbl">Emails Sent</div></div>
  <div class="card"><div class="val">{{.Opened}} <small style="font-size:16px;color:#636e72">{{formatRate .OpenRate}}</small></div><div class="lbl">Opened</div></div>
  <div class="card"><div class="val">{{.Clicked}} <small style="font-size:16px;color:#636e72">{{formatRate .ClickRate}}</small></div><div class="lbl">Clicked</div></div>
  <div class="card"><div class="val">{{.Submitted}} <small style="font-size:16px;color:#636e72">{{formatRate .SubmissionRate}}</small></div><div class="lbl">Submitted Creds</div></div>
</div>

<!-- Per-target breakdown -->
<h2>👤 Target Breakdown</h2>
<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Email</th>
      <th>Sent</th>
      <th>Opened</th>
      <th>Clicked</th>
      <th>Submitted</th>
      <th>Risk</th>
    </tr>
  </thead>
  <tbody>
    {{range .Targets}}
    <tr>
      <td>{{.FirstName}} {{.LastName}}</td>
      <td>{{.Email}}</td>
      <td class="{{if .EmailSent}}check-yes{{else}}check-no{{end}}">{{checkmark .EmailSent}}</td>
      <td class="{{if .Opened}}check-yes{{else}}check-no{{end}}">{{checkmark .Opened}}</td>
      <td class="{{if .Clicked}}check-yes{{else}}check-no{{end}}">{{checkmark .Clicked}}</td>
      <td class="{{if .Submitted}}check-yes{{else}}check-no{{end}}">{{checkmark .Submitted}}</td>
      <td>{{riskBadge .RiskLevel}}</td>
    </tr>
    {{end}}
  </tbody>
</table>

<!-- Timeline -->
<h2>🕒 Event Timeline</h2>
{{range .Timeline}}
<div class="timeline-entry">
  <div class="timeline-time">{{formatTime .Timestamp}}</div>
  <div class="timeline-event">{{eventLabel .EventType}}</div>
  <div class="timeline-who">{{.FirstName}} {{.LastName}} &lt;{{.Email}}&gt;</div>
  <div class="timeline-ip">{{.IP}}</div>
</div>
{{end}}

</body>
</html>`))

func (r *Reporter) GenerateHTML(ctx context.Context, campaignID string) ([]byte, error) {
	report, err := r.Build(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := reportTmpl.Execute(&buf, report); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Reporter) SaveHTML(ctx context.Context, campaignID, path string) error {
	html, err := r.GenerateHTML(ctx, campaignID)
	if err != nil {
		return err
	}
	return os.WriteFile(path, html, 0644)
}
