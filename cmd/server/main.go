package main

import (
	"log"
	"net/http"
	"os"
	inboundhttp "promotion-service/internal/adapters/inbound/http"
	"promotion-service/internal/adapters/outbound/memory"
	"promotion-service/internal/adapters/outbound/postgres"
	"promotion-service/internal/application"
)

func main() {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		pgRepo, err := postgres.NewPromotionRepository(databaseURL)
		if err != nil {
			log.Fatalf("cannot connect postgres: %v", err)
		}
		defer func() {
			_ = pgRepo.Close()
		}()

		service := application.NewPromotionService(pgRepo)
		handler := inboundhttp.NewHandler(service)

		mux := http.NewServeMux()
		handler.Register(mux)

		log.Println("promotion service listening on :8080 (postgres)")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatal(err)
		}
		return
	}

	repo := memory.NewPromotionRepository()
	service := application.NewPromotionService(repo)
	handler := inboundhttp.NewHandler(service)

	mux := http.NewServeMux()
	handler.Register(mux)

	log.Println("promotion service listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
