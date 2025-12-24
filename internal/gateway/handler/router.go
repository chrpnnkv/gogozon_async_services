package handler

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Router struct {
	ordersProxy   *httputil.ReverseProxy
	paymentsProxy *httputil.ReverseProxy
}

func NewRouter(ordersBase, paymentsBase *url.URL) *Router {
	return &Router{
		ordersProxy:   newReverseProxy(ordersBase),
		paymentsProxy: newReverseProxy(paymentsBase),
	}
}

func (rt *Router) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/orders", rt.ordersProxy.ServeHTTP)
	mux.HandleFunc("/orders/", rt.ordersProxy.ServeHTTP)
	mux.HandleFunc("/accounts", rt.paymentsProxy.ServeHTTP)
	mux.HandleFunc("/accounts/", rt.paymentsProxy.ServeHTTP)
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = target.Host
		if req.Header.Get("X-Forwarded-Host") == "" {
			req.Header.Set("X-Forwarded-Host", req.Host)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[gateway] proxy error: %v", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return proxy
}
