// cmd/server/main.go
package main

import (
	"log"
	"net/http"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/admin"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/mailer"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/reporter"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/tracker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	database, err := db.New("phishsim.db")
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer database.Close()
	log.Println("DB ready ✓")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Tracker (public)
	t := tracker.New(database)
	r.Mount("/t", t.Routes())

	// Admin dashboard
	rep := reporter.New(database)
	m := mailer.New(mailer.Config{
		Host:     "sandbox.smtp.mailtrap.io",
		Port:     587,
		Username: "126178ed9fa520",
		Password: "d561589c1d6378",
		BaseURL:  "http://localhost:8080",
	}, database)
	a := admin.New(database, rep, m)
	r.Mount("/admin", a.Routes())

	// Redirect root to admin
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
