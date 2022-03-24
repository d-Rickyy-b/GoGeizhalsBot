package prometheus

import (
	"GoGeizhalsBot/internal/database"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
)

var (
	totalUniqueUsers = metrics.NewGauge("gogeizhalsbot_unique_users_total", func() float64 {
		return float64(database.GetUserCount())
	})
	totalUniquePriceagents = metrics.NewGauge("gogeizhalsbot_unique_priceagents", func() float64 {
		return float64(database.GetPriceAgentCount())
	})
	totalUniqueProductPriceagents = metrics.NewGauge("gogeizhalsbot_unique_priceagents{type=\"product\"}", func() float64 {
		return float64(database.GetPriceAgentProductCount())
	})
	totalUniqueWishlistPriceagents = metrics.NewGauge("gogeizhalsbot_unique_priceagents{type=\"wishlist\"}", func() float64 {
		return float64(database.GetPriceAgentWishlistCount())
	})
	totalUniqueProductPriceagentsValue int64

	TotalUserInteractions   = metrics.NewCounter("gogeizhalsbot_user_interactions_total")
	GeizhalsHTTPRequests    = metrics.NewCounter("gogeizhalsbot_geizhals_http_requests_total")
	PriceagentNotifications = metrics.NewCounter("gogeizhalsbot_priceagent_notifications_total")
	ProxyErrors             = metrics.NewCounter("gogeizhalsbot_proxy_errors_total")
	GraphsRendered          = metrics.NewCounter("gogeizhalsbot_graphs_rendered_total")
)

// var backgroundUpdateChecks = metrics.NewSummary("gogeizhalsbot_total_requests")

func StartPrometheusExporter(addr string) {
	// Expose the registered metrics at `/metrics` path.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, true)
	})
	http.ListenAndServe(addr, nil)
}
