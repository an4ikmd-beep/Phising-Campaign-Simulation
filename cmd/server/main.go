package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/tracker"
)

func main() {
	database, err := db.New("phishsim.db")

	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer database.Close()

	log.Println("DB READY")

	t := tracker.New(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/t", t.Routes())
	log.Println("Tracker running on http://localhost:8080")
	http.ListenAndServe(":8080", r)

}
