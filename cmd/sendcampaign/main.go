package main

import (
	"context"
	"log"
	"os"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/mailer"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/sendcampaign <campaign-id>")
	}
	campaignID := os.Args[1]

	database, err := db.New("phishsim.db")
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	m := mailer.New(mailer.Config{
		Host:     "sandbox.smtp.mailtrap.io",
		Port:     587,
		Username: "126178ed9fa520", // ← paste from Mailtrap
		Password: "d561589c1d6378", // ← paste from Mailtrap
		BaseURL:  "http://localhost:8080",
	}, database)

	log.Printf("Sending campaign %s ...", campaignID)
	if err := m.SendCampaign(context.Background(), campaignID); err != nil {
		log.Fatalf("send: %v", err)
	}
	log.Println("Done!")
}
