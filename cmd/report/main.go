package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/db"
	"github.com/an4ikmd-beep/Phising-Campaign-Simulation/internal/reporter"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/report <campaign-id>")
	}
	campaignID := os.Args[1]

	database, err := db.New("phishsim.db")
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	r := reporter.New(database)
	outFile := fmt.Sprintf("report-%s.html", campaignID[:8])

	if err := r.SaveHTML(context.Background(), campaignID, outFile); err != nil {
		log.Fatalf("generate report: %v", err)
	}

	log.Printf("✓ Report saved to %s", outFile)
}
