package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/mailer"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/reporter"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	DB       *db.DB
	Reporter *reporter.Reporter
	Mailer   *mailer.Mailer
	Tmpls    *template.Template
}

func New(database *db.DB, r *reporter.Reporter, m *mailer.Mailer) *Handler {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
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
				"Critical": "#c0392b", "High": "#e67e22",
				"Medium": "#f1c40f", "Low": "#27ae60",
			}
			return template.HTML(fmt.Sprintf(
				`<span style="background:%s;color:white;padding:2px 8px;border-radius:4px;font-size:12px;">%s</span>`,
				colors[s], s,
			))
		},
		"checkmark": func(b bool) string {
			if b {
				return "✓"
			}
			return "—"
		},
	}

	tmpls := template.Must(
		template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"),
	)

	return &Handler{DB: database, Reporter: r, Mailer: m, Tmpls: tmpls}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listCampaigns)
	r.Get("/new", h.newCampaignForm)
	r.Post("/new", h.createCampaign)
	r.Get("/campaign/{id}", h.viewCampaign)
	r.Post("/campaign/{id}/send", h.sendCampaign)
	r.Get("/campaign/{id}/report", h.downloadReport)
	return r
}

// ---------- Handlers ----------

type campaignRow struct {
	Campaign db.Campaign
	Counts   struct {
		EmailSent int
		Opened    int
		Clicked   int
		Submitted int
	}
}

func (h *Handler) listCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.DB.ListCampaigns(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var rows []campaignRow
	for _, c := range campaigns {
		counts, _ := h.DB.CountByType(r.Context(), c.ID)
		row := campaignRow{Campaign: c}
		row.Counts.EmailSent = counts["email_sent"]
		row.Counts.Opened = counts["opened"]
		row.Counts.Clicked = counts["clicked"]
		row.Counts.Submitted = counts["submitted"]
		rows = append(rows, row)
	}

	h.Tmpls.ExecuteTemplate(w, "campaigns", map[string]any{
		"Campaigns": rows,
	})
}

func (h *Handler) newCampaignForm(w http.ResponseWriter, r *http.Request) {
	h.Tmpls.ExecuteTemplate(w, "new", nil)
}

func (h *Handler) createCampaign(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	campaign := db.Campaign{
		Name:        r.FormValue("name"),
		Subject:     r.FormValue("subject"),
		SenderName:  r.FormValue("sender_name"),
		SenderEmail: r.FormValue("sender_email"),
		Status:      "active",
	}

	if err := h.DB.CreateCampaign(r.Context(), campaign); err != nil {
		h.Tmpls.ExecuteTemplate(w, "new", map[string]any{"Error": err.Error()})
		return
	}

	// fetch it back to get the ID
	campaigns, _ := h.DB.ListCampaigns(r.Context())
	newCampaign := campaigns[0]

	// parse targets: one per line, format: email,firstname,lastname
	for _, line := range strings.Split(r.FormValue("targets"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 1 {
			continue
		}
		t := db.Target{
			CampaignID: newCampaign.ID,
			Email:      strings.TrimSpace(parts[0]),
		}
		if len(parts) > 1 {
			t.FirstName = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			t.LastName = strings.TrimSpace(parts[2])
		}
		h.DB.CreateTarget(r.Context(), t)
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) viewCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	report, err := h.Reporter.Build(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.Tmpls.ExecuteTemplate(w, "campaign", map[string]any{
		"CampaignID": id,
		"Report":     report,
	})
}

func (h *Handler) sendCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.Mailer.SendCampaign(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send: %v", err), 500)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/campaign/%s", id), http.StatusSeeOther)
}

func (h *Handler) downloadReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	html, err := h.Reporter.GenerateHTML(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="report-%s.html"`, id[:8]))
	w.Write(html)
}
