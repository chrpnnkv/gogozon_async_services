package config

import (
	"log"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	OrdersBaseURL   *url.URL
	PaymentsBaseURL *url.URL
}

func MustLoad() *Config {
	return &Config{
		OrdersBaseURL:   mustURL("ORDERS_BASE_URL"),
		PaymentsBaseURL: mustURL("PAYMENTS_BASE_URL"),
	}
}

func mustURL(envKey string) *url.URL {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		log.Fatalf("missing env %s", envKey)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		log.Fatalf("bad %s=%q: expected like http://host:port", envKey, raw)
	}
	return u
}
