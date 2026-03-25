package tracker

import (
	"net/http"
	"strings"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/go-chi/chi/v5"
)

type Tracker struct {
	DB *db.DB
}

func New(database *db.DB) *Tracker {
	return &Tracker{DB: database}
}

func (t *Tracker) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/open/{token}.png", t.handleOpen)
	r.Get("/click/{token}", t.handleClick)
	r.Get("/page/{token}", t.handlePage)
	r.Post("/submit/{token}", t.handleSubmit)
	return r
}

// 1x1 transparent GIF
var pixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff,
	0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

func (t *Tracker) handleOpen(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	target, err := t.DB.GetTargetByToken(r.Context(), token)
	if err != nil {
		//still return pixel if token unknown
		w.Header().Set("Content-Type", "image/gif")
		w.Write(pixel)
		return
	}

	t.DB.LogEvent(r.Context(), db.Event{
		CampaignID: target.CampaignID,
		TargetID:   target.ID,
		EventType:  "opened",
		IP:         getIP(r),
		UserAgent:  r.UserAgent(),
	})

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(pixel)
}

func (t *Tracker) handleClick(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	target, err := t.DB.GetTargetByToken(r.Context(), token)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	t.DB.LogEvent(r.Context(), db.Event{
		CampaignID: target.CampaignID,
		TargetID:   target.ID,
		EventType:  "clicked",
		IP:         getIP(r),
		UserAgent:  r.UserAgent(),
	})

	//redir to fake login page
	http.Redirect(w, r, "/t/page/"+token, http.StatusFound)
}

func (t *Tracker) handlePage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	// Fake Microsoft login page
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Sign in to your account</title>
<style>
  body { font-family: Segoe UI, sans-serif; display:flex; justify-content:center; padding-top:80px; background:#f3f3f3; }
  .box { background:white; padding:44px; width:380px; box-shadow:0 2px 6px rgba(0,0,0,.1); }
  h1 { font-size:24px; font-weight:600; margin-bottom:16px; }
  input { width:100%; padding:8px; margin:8px 0 16px; border:1px solid #ccc; font-size:15px; box-sizing:border-box; }
  button { width:100%; padding:10px; background:#0067b8; color:white; border:none; font-size:15px; cursor:pointer; }
</style>
</head>
<body>
<div class="box">
  <h1>Sign in</h1>
  <form method="POST" action="/t/submit/` + token + `">
    <input name="email" type="email" placeholder="Email" required />
    <input name="password" type="password" placeholder="Password" required />
    <button type="submit">Sign in</button>
  </form>
</div>
</body>
</html>`))
}

func (t *Tracker) handleSubmit(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	r.ParseForm()
	// we log that submission happened, but do NOT store actual credentials
	_ = r.FormValue("email")

	target, err := t.DB.GetTargetByToken(r.Context(), token)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	t.DB.LogEvent(r.Context(), db.Event{
		CampaignID: target.CampaignID,
		TargetID:   target.ID,
		EventType:  "submitted",
		IP:         getIP(r),
		UserAgent:  r.UserAgent(),
	})

	// Show awareness page instead of stealing creds
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Security Awareness Training</title>
<style>
  body { font-family: Segoe UI, sans-serif; display:flex; justify-content:center; padding-top:80px; background:#f3f3f3; }
  .box { background:white; padding:44px; width:420px; box-shadow:0 2px 6px rgba(0,0,0,.1); text-align:center; }
  h1 { color:#d83b01; }
  p { color:#444; line-height:1.6; }
</style>
</head>
<body>
<div class="box">
  <h1>⚠️ This was a phishing simulation</h1>
  <p>You entered your credentials on a fake login page.<br>
  In a real attack, your account would now be compromised.</p>
  <p><strong>Tips:</strong> Always check the URL before signing in.
  Enable multi-factor authentication on all accounts.</p>
</div>
</body>
</html>`))
}

func getIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}
