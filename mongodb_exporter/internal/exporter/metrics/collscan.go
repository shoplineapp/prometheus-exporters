package metrics

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"mongodb_performance_exporter/interfaces"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type CollectionScanMetric struct {
	interfaces.MetricProcesser

	logger *logrus.Entry

	queries []CollectionScanQuery
	buffer  []CollectionScanQuery
}

type CollectionScanQuery struct {
	Cluster      string
	Server       string
	Raw          string  `json:"line_str"`
	Namespace    string  `json:"namespace"`
	Operation    string  `json:"operation"`
	NReturned    int32   `json:"nreturned"`
	NScanned     int32   `json:"nscanned"`
	NumYields    int32   `json:"numYields"`
	Duration     float32 `json:"duration"`
	DocsExamined int32
}

type CollectionScanCounter struct {
	Cluster   string
	Server    string
	Namespace string
	Operation string
	Count     int32
}

var metricLabelNames = []string{"cluster", "server", "namespace", "operation", "raw"}

func (q CollectionScanQuery) Labels() []string {
	return []string{
		q.Cluster,
		q.Server,
		q.Namespace,
		q.Operation,
		q.Raw,
	}
}

func (c CollectionScanCounter) Labels() []string {
	return []string{
		c.Cluster,
		c.Server,
		c.Namespace,
		c.Operation,
	}
}

var (
	MongoAtlasCollectionScanCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_collscan_queries_count",
			Help: "The count of collection scan queries from logs of mongodb atlas database",
		},
		[]string{"cluster", "server", "namespace", "operation"},
	)

	MongoAtlasCollectionScanDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_collscan_queries_duration",
			Help: "The duration of collection scan queries from logs of mongodb atlas database",
		},
		metricLabelNames,
	)

	MongoAtlasCollectionScanNScanned = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_collscan_queries_nscanned",
			Help: "The nscanned count of collection scan queries from logs of mongodb atlas database",
		},
		metricLabelNames,
	)

	MongoAtlasCollectionScanDocsExaminated = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mongo_atlas_collscan_queries_docs_examinated",
			Help: "The docsExaminated count of collection scan queries from logs of mongodb atlas database",
		},
		metricLabelNames,
	)
)

func (m *CollectionScanMetric) InitMetrics() {
	m.logger.Error("Metric initialized")
}

func (m *CollectionScanMetric) ParseFile(file *os.File, cluster string, server string) {
	var out bytes.Buffer
	var stderrr bytes.Buffer

	// mtools disallow call from non-TTY context and throw an error with "this tool can't parse input from stdin"
	// https://github.com/rueckstiess/mtools/issues/404
	// Skipping cluster info and table headers
	cmd := exec.Command("script", "-q", "-c", fmt.Sprintf("mlogfilter --planSummary COLLSCAN --json %s", file.Name()))
	cmd.Stdout = &out
	cmd.Stderr = &stderrr
	err := cmd.Run()
	if err != nil {
		m.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": err}).Error("Error parsing logs")
	}

	entries := []CollectionScanQuery{}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		line := scanner.Text()

		query := CollectionScanQuery{
			Cluster: cluster,
			Server:  server,
		}
		json.Unmarshal([]byte(line), &query)

		if query.Duration <= 0 {
			continue
		}

		// parse document examined count
		re, _ := regexp.Compile(`docsExamined: *(\d+) *`)
		match := re.FindStringSubmatch(query.Raw)
		if len(match) >= 2 {
			val, _ := strconv.Atoi(match[1])
			query.DocsExamined = int32(val)
		}
		entries = append(entries, query)
	}

	m.buffer = append(m.buffer, entries...)
	m.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(entries)}).Debug("Log entries saved into buffer")
}

func (m *CollectionScanMetric) UpdateMetrics() {
	m.logger.WithFields(logrus.Fields{"buffer": len(m.buffer)}).Debug("Flushing buffer to store")

	for _, entry := range m.queries {
		MongoAtlasCollectionScanDuration.DeleteLabelValues(entry.Labels()...)
		MongoAtlasCollectionScanNScanned.DeleteLabelValues(entry.Labels()...)
		MongoAtlasCollectionScanDocsExaminated.DeleteLabelValues(entry.Labels()...)
		MongoAtlasCollectionScanCount.DeleteLabelValues(entry.Cluster, entry.Server, entry.Namespace, entry.Operation)
	}

	m.queries = m.buffer

	m.logger.WithFields(logrus.Fields{"count": len(m.queries)}).Info("Reporting data to metric API")
	count := map[string]*CollectionScanCounter{}
	for _, entry := range m.queries {
		MongoAtlasCollectionScanDuration.WithLabelValues(entry.Labels()...).Set(float64(entry.Duration))
		if entry.NScanned > 0 {
			MongoAtlasCollectionScanNScanned.WithLabelValues(entry.Labels()...).Set(float64(entry.NScanned))
		}
		// if entry.DocsExamined > 0 {
		MongoAtlasCollectionScanDocsExaminated.WithLabelValues(entry.Labels()...).Set(float64(entry.DocsExamined))
		// }

		key := fmt.Sprintf("%s.%s.%s.%s", entry.Cluster, entry.Server, entry.Namespace, entry.Operation)
		if count[key] == nil {
			count[key] = &CollectionScanCounter{
				Cluster:   entry.Cluster,
				Server:    entry.Server,
				Namespace: entry.Namespace,
				Operation: entry.Operation,
				Count:     0,
			}
		}
		count[key].Count++
	}

	// group by namespace and operation and then count
	for _, counter := range count {
		MongoAtlasCollectionScanCount.WithLabelValues(counter.Labels()...).Set(float64(counter.Count))
	}
	m.buffer = []CollectionScanQuery{}
}

func NewCollectionScanMetric(logger *logger.Logger) *CollectionScanMetric {
	m := &CollectionScanMetric{
		logger: logger.WithFields(logrus.Fields{"metric": "CollectionScanMetric"}),
	}
	return m
}
