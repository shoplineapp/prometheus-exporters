package exporter

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metric struct {
	store *Store
}

func (m Metric) InitMetrics() {
	m.store.InitMetrics()
}

func (m Metric) Serve() {
	router := gin.Default()
	m.InitMetrics()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Run(":3000")
}

func NewMetric(store *Store) *Metric {
	return &Metric{
		store: store,
	}
}
