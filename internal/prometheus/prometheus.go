package prometheus

import (
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

var (
	TotalUniqueUsersValue int64
	totalUniqueUsers      = metrics.NewGauge("gogeizhalsbot_unique_users_total", func() float64 {
		return float64(TotalUniqueUsersValue)
	})

	TotalUniquePriceagentsValue int64
	totalUniquePriceagents      = metrics.NewGauge("gogeizhalsbot_unique_priceagents", func() float64 {
		return float64(TotalUniquePriceagentsValue)
	})

	TotalUniqueProductPriceagentsValue int64
	totalUniqueProductPriceagents      = metrics.NewGauge("gogeizhalsbot_unique_priceagents{type=\"product\"}", func() float64 {
		return float64(TotalUniqueProductPriceagentsValue)
	})

	TotalUniqueWishlistPriceagentsValue int64
	totalUniqueWishlistPriceagents      = metrics.NewGauge("gogeizhalsbot_unique_priceagents{type=\"wishlist\"}", func() float64 {
		return float64(TotalUniqueWishlistPriceagentsValue)
	})

	TotalUserInteractions   = metrics.NewCounter("gogeizhalsbot_user_interactions_total")
	GeizhalsHTTPRequests    = metrics.NewCounter("gogeizhalsbot_geizhals_http_requests_total")
	HTTPRequests429         = metrics.NewCounter("gogeizhalsbot_geizhals_http_requests_429_total")
	PriceagentNotifications = metrics.NewCounter("gogeizhalsbot_priceagent_notifications_total")
	HTTPErrors              = metrics.NewCounter("gogeizhalsbot_http_errors_total")
	GraphsRendered          = metrics.NewCounter("gogeizhalsbot_graphs_rendered_total")
)

// var backgroundUpdateChecks = metrics.NewSummary("gogeizhalsbot_total_requests")

func StartPrometheusExporter(addr string) error {
	// Expose the registered metrics at `/metrics` path.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, true)
	})

	server := &http.Server{
		Addr:              addr,
		IdleTimeout:       time.Minute,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	return server.ListenAndServe()
}
