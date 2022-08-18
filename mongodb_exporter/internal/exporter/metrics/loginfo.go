package metrics

import (
	"bufio"
	"bytes"
	"fmt"
	"mongodb_performance_exporter/interfaces"
	"mongodb_performance_exporter/internal/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type LogInfoMetric struct {
	interfaces.MetricProcesser

	logger *logger.Logger

	queries []LogInfoQuery
	buffer  []LogInfoQuery
}

type LogInfoQuery struct {
	Cluster   string
	Server    string
	Namespace string
	Operation string
	Pattern   string
	Count     int32
	MinMS     float32
	MaxMS     float32
	P95MS     float32
	SumMS     float32
	MeanMS    float32
}

func (q LogInfoQuery) Labels() []string {
	return []string{
		q.Cluster,
		q.Server,
		q.Namespace,
		q.Operation,
		q.Pattern,
	}
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

func (m *LogInfoMetric) InitMetrics() {
	m.logger.WithFields(logrus.Fields{"metric": "LogInfoMetric"}).Error("Metric initialized")
}

func (m *LogInfoMetric) ParseFile(file *os.File, cluster string, server string) {
	var out bytes.Buffer
	var stderrr bytes.Buffer

	// mtools disallow call from non-TTY context and throw an error with "this tool can't parse input from stdin"
	// https://github.com/rueckstiess/mtools/issues/404
	// Skipping cluster info and table headers
	cmd := exec.Command("script", "-q", "-c", fmt.Sprintf("mloginfo --no-progressbar --queries %s | tail -n +15", file.Name()))
	cmd.Stdout = &out
	cmd.Stderr = &stderrr
	err := cmd.Run()
	if err != nil {
		m.logger.WithFields(logrus.Fields{"metric": "LogInfoMetric", "cluster": cluster, "server": server, "error": err}).Error("Error parsing logs")
	}

	entries := []LogInfoQuery{}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		line := scanner.Text()

		// Columns are turned from tab into multiple spaces from stdout return
		parts := utils.RegSplit(line, "[ ]{3,}")
		if len(parts) < 8 {
			continue
		}
		entries = append(entries, LogInfoQuery{
			Cluster:   cluster,
			Server:    server,
			Namespace: parts[0],
			Operation: parts[1],
			Pattern:   parts[2],
			Count:     utils.StrToInt32(parts[3]),
			MinMS:     utils.StrToFloat32(parts[4]),
			MaxMS:     utils.StrToFloat32(parts[5]),
			P95MS:     utils.StrToFloat32(parts[6]),
			SumMS:     utils.StrToFloat32(parts[7]),
			MeanMS:    utils.StrToFloat32(parts[8]),
		})
	}

	m.buffer = append(m.buffer, entries...)
	m.logger.WithFields(logrus.Fields{"metric": "LogInfoMetric", "cluster": cluster, "server": server, "count": len(entries)}).Debug("Log entries saved into buffer")
}

func (m *LogInfoMetric) UpdateMetrics() {
	m.logger.WithFields(logrus.Fields{"metric": "LogInfoMetric", "buffer": len(m.buffer)}).Debug("Flushing buffer to store")

	for _, entry := range m.queries {
		MongoAtlasTopQueryCount.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMinMS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMaxMS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryP95MS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMeanMS.DeleteLabelValues(entry.Labels()...)
	}

	m.queries = m.buffer

	m.logger.WithFields(logrus.Fields{"metric": "LogInfoMetric", "count": len(m.queries)}).Info("Reporting data to metric API")
	for _, entry := range m.queries {
		MongoAtlasTopQueryCount.WithLabelValues(entry.Labels()...).Set(float64(entry.Count))
		MongoAtlasTopQueryMinMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MinMS))
		MongoAtlasTopQueryMaxMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MaxMS))
		MongoAtlasTopQueryP95MS.WithLabelValues(entry.Labels()...).Set(float64(entry.P95MS))
		MongoAtlasTopQueryMeanMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MeanMS))
	}
	m.buffer = []LogInfoQuery{}
}

func NewLogInfoMetric(logger *logger.Logger) *LogInfoMetric {
	p := &LogInfoMetric{
		logger: logger,
	}
	return p
}
