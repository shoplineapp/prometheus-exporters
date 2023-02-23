package metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mongodb_performance_exporter/interfaces"
	"os"
	"regexp"

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
	Cluster   string
	Server    string
	Raw       string
	Namespace string
	Operation string
	Attr      struct {
		DurationMillis int32                  `json:"durationMillis"`
		DocsExamined   int32                  `json:"docsExamined"`
		NumYields      int32                  `json:"numYields"`
		NReturned      int32                  `json:"nreturned"`
		Namespace      string                 `json:"ns"`
		Command        map[string]interface{} `json:"command"`
	}
}

type CollectionScanCounter struct {
	Cluster   string
	Server    string
	Namespace string
	Operation string
	Count     int32
}

var metricLabelNames = []string{"cluster", "server", "namespace", "operation", "raw"}

var (
	COMMAND_OPERATIONS = []string{"query", "getMore", "update", "remove", "count", "findAndModify", "geoNear", "find", "aggregate", "command"}
)

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

func (m *CollectionScanMetric) ParseFile(legacyFile *os.File, originalFile *os.File, cluster string, server string) {
	f, err := os.Open(originalFile.Name())
	if err != nil {
		panic(err)
	}
	defer f.Close()

	entries := []CollectionScanQuery{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		re, _ := regexp.Compile(`.+\"planSummary\":\"COLLSCAN\".+`)
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}

		query := CollectionScanQuery{
			Cluster: cluster,
			Server:  server,
			Raw:     line,
		}
		json.Unmarshal([]byte(line), &query)

		if query.Attr.DurationMillis <= 0 {
			continue
		}
		query.Namespace = query.Attr.Namespace
		query.Operation = m.GetOperation(query.Attr.Command)

		entries = append(entries, query)
	}

	m.buffer = append(m.buffer, entries...)
	m.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(entries)}).Debug("Log entries saved into buffer")
}

func (m *CollectionScanMetric) UpdateMetrics() {
	m.logger.WithFields(logrus.Fields{"buffer": len(m.buffer)}).Debug("Flushing buffer to store")

	MongoAtlasCollectionScanDuration.Reset()
	MongoAtlasCollectionScanDocsExaminated.Reset()
	MongoAtlasCollectionScanCount.Reset()
	m.queries = m.buffer

	m.logger.WithFields(logrus.Fields{"count": len(m.queries)}).Info("Reporting data to metric API")
	count := map[string]*CollectionScanCounter{}
	for _, entry := range m.queries {
		MongoAtlasCollectionScanDuration.WithLabelValues(entry.Labels()...).Set(float64(entry.Attr.DurationMillis))
		if entry.Attr.DocsExamined > 0 {
			MongoAtlasCollectionScanDocsExaminated.WithLabelValues(entry.Labels()...).Set(float64(entry.Attr.DocsExamined))
		}
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

func (m *CollectionScanMetric) GetOperation(attr map[string]interface{}) string {
	for _, cmd := range COMMAND_OPERATIONS {
		if _, ok := attr[cmd]; ok {
			return cmd
		}
	}
	return ""
}

func NewCollectionScanMetric(logger *logger.Logger) *CollectionScanMetric {
	m := &CollectionScanMetric{
		logger: logger.WithFields(logrus.Fields{"metric": "CollectionScanMetric"}),
	}
	return m
}
