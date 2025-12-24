package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"HW4/internal/common/kafka"
	"HW4/internal/orders/handler"
	"HW4/internal/orders/repository"
	"HW4/internal/orders/service"
	"HW4/internal/orders/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("postgres", mustEnv("ORDERS_DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	brokers := kafka.SplitBrokers(mustEnv("KAFKA_BROKERS"))

	producer := kafka.NewProducer(brokers)
	defer producer.Close()
	go worker.NewOutboxPublisher(db, producer).Run(ctx)

	resTopic := mustEnv("KAFKA_TOPIC_PAYMENT_RESULT")
	group := mustEnv("ORDERS_CONSUMER_GROUP")
	reqTopic := mustEnv("KAFKA_TOPIC_PAYMENT_REQUESTED")

	log.Printf("[orders] starting consumer topic=%s group=%s brokers=%v", resTopic, group, brokers)

	resConsumer := kafka.NewConsumer(brokers, resTopic, group)
	defer resConsumer.Close()

	statusRepo := repository.NewOrdersStatusRepo(db)
	go worker.NewPaymentResultConsumer(resConsumer, statusRepo).Run(ctx)

	ordersRepo := repository.NewOrdersRepo(db, reqTopic)
	ordersSvc := service.New(ordersRepo)
	h := handler.New(ordersSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateOrder(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.ListOrders(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetOrder(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() { <-ctx.Done(); _ = srv.Shutdown(context.Background()) }()

	log.Println("[orders] up on :8080")
	log.Fatal(srv.ListenAndServe())
}

func mustEnv(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}
