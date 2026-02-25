package main

import (
	"log"
	"net/http"
	inboundhttp "promotion-service/internal/adapters/inbound/http"
	"promotion-service/internal/adapters/outbound/memory"
	"promotion-service/internal/application"
)

func main() {
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
