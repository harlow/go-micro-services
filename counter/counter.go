package counter

import "github.com/prometheus/client_golang/prometheus"

// expose metrics
// go func() {
// 	http.Handle("/metrics", promhttp.Handler())
// 	log.Fatal(http.ListenAndServe(":6060", nil))
// }()

func init() {
	// gmailScans.WithLabelValues(statusError).Inc()
	// prometheus.MustRegister(gmailScans)
}

// New returns a new counter
func New(name string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        name,
			Help:        "Number of mailboxes scanned.",
			ConstLabels: prometheus.Labels{},
		},
		[]string{"status"},
	)
}
