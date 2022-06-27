package exporter

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metric struct {
}

var (
	MongoAtlasTopQueryCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_top_queries_count",
			Help: "The count of top queries result from mloginfo of mongodb atlas database",
		},
		[]string{
			"cluster",
			"server",
			"namespace",
			"operation",
			"pattern",
		},
	)

	MongoAtlasTopQueryMinMS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_top_queries_duration_ms_min",
			Help: "The min duration of top queries result from mloginfo of mongodb atlas database",
		},
		[]string{
			"cluster",
			"server",
			"namespace",
			"operation",
			"pattern",
		},
	)

	MongoAtlasTopQueryMaxMS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_top_queries_duration_ms_max",
			Help: "The max duration of top queries result from mloginfo of mongodb atlas database",
		},
		[]string{
			"cluster",
			"server",
			"namespace",
			"operation",
			"pattern",
		},
	)

	MongoAtlasTopQueryP95MS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_top_queries_duration_ms_p95",
			Help: "The p95 duration of top queries result from mloginfo of mongodb atlas database",
		},
		[]string{
			"cluster",
			"server",
			"namespace",
			"operation",
			"pattern",
		},
	)

	MongoAtlasTopQueryMeanMS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_top_queries_duration_ms_mean",
			Help: "The mean duration of top queries result from mloginfo of mongodb atlas database",
		},
		[]string{
			"cluster",
			"server",
			"namespace",
			"operation",
			"pattern",
		},
	)
)

func (m Metric) InitMetrics() {

}

func (m Metric) Serve() {
	router := gin.Default()
	m.InitMetrics()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Run(":3000")
}

func NewMetric() *Metric {
	return &Metric{}
}
