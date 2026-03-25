package main

import (
	"context"
	"fmt"
	"log"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
)

func main() {
	database, err := db.New("phishsim.db")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	campaign := db.Campaign{
		Name:        "Test Campaign #1",
		Subject:     "Action Required: Verify your account",
		SenderName:  "It Security Team",
		SenderEmail: "security@company-internal.com",
		Status:      "active",
	}

	err = database.CreateCampaign(ctx, campaign)
	if err != nil {
		log.Fatalf("create campaign: %v", err)
	}
	fmt.Println("Campaign created")

	campaigns, err := database.ListCampaigns(ctx)
	if err != nil {
		log.Fatalf("list campaigns: %v", err)
	}
	c := campaigns[0]
	fmt.Printf("\tID: %s\n", c.ID)

	targets := []db.Target{
		{CampaignID: c.ID, Email: "alice@testorg.com", FirstName: "Alice", LastName: "Smith"},
		{CampaignID: c.ID, Email: "bob@testorg.com", FirstName: "Bob", LastName: "Jones"},
		{CampaignID: c.ID, Email: "carol@testorg.com", FirstName: "Carol", LastName: "White"},
	}

	for _, t := range targets {
		if err := database.CreateTarget(ctx, t); err != nil {
			log.Fatalf("create target %s: %v", t.Email, err)
		}
	}
	fmt.Println("targets created")

	allTargets, err := database.GetTargetsByCampaign(ctx, c.ID)
	if err != nil {
		log.Fatalf("get targets: %v", err)
	}

	fmt.Println("\n--- Tracking URLs (open in browser to test) ---")
	for _, t := range allTargets {
		fmt.Printf("\n%s %s (%s)\n", t.FirstName, t.LastName, t.Email)
		fmt.Printf("  Open pixel : http://localhost:8080/t/open/%s.png\n", t.Token)
		fmt.Printf("  Click link : http://localhost:8080/t/click/%s\n", t.Token)
		fmt.Printf("  Direct page: http://localhost:8080/t/page/%s\n", t.Token)
	}

	fmt.Println("\nSeed complete — run the server then visit the URLs above")
}
