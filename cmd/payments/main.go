package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"HW4/internal/common/kafka"
	"HW4/internal/payments/handler"
	"HW4/internal/payments/repository"
	"HW4/internal/payments/service"
	"HW4/internal/payments/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("postgres", mustEnv("PAYMENTS_DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	accRepo := repository.NewAccountsRepo(db)
	paySvc := service.New(accRepo)
	h := handler.New(paySvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateAccount(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/accounts/topup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.TopUp(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetBalance(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	brokers := kafka.SplitBrokers(mustEnv("KAFKA_BROKERS"))
	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	go worker.NewOutboxPublisher(db, producer).Run(ctx)

	reqTopic := mustEnv("KAFKA_TOPIC_PAYMENT_REQUESTED")
	group := mustEnv("PAYMENTS_CONSUMER_GROUP")

	consumer := kafka.NewConsumer(brokers, reqTopic, group)
	defer consumer.Close()

	resultTopic := mustEnv("KAFKA_TOPIC_PAYMENT_RESULT")
	processor := repository.NewPaymentProcessor(db, resultTopic)
	go worker.NewPaymentRequestedConsumer(consumer, processor, producer).Run(ctx)

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() { <-ctx.Done(); _ = srv.Shutdown(context.Background()) }()

	log.Println("[payments] up on :8080")
	log.Fatal(srv.ListenAndServe())
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}
