package fcm

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	pSentMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fcm_sent_messages",
		Help: "Number of messages sent to FCM",
	})
)

func init() {
	prometheus.MustRegister(
		pSentMessages,
	)
}
