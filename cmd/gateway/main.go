package main

import (
	"log"
	"net/http"
	"time"

	"HW4/internal/gateway/config"
	"HW4/internal/gateway/handler"
)

func main() {
	cfg := config.MustLoad()
	mux := http.NewServeMux()
	rt := handler.NewRouter(cfg.OrdersBaseURL, cfg.PaymentsBaseURL)
	rt.Register(mux)
	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Println("[gateway] up on :8080")
	log.Fatal(srv.ListenAndServe())
}
