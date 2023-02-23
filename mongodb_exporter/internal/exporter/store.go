package exporter

import (
	"fmt"
	"mongodb_performance_exporter/interfaces"
	metrics "mongodb_performance_exporter/internal/exporter/metrics"
	"os"
	"sync"

	"github.com/shoplineapp/go-app/plugins/env"
	"github.com/shoplineapp/go-app/plugins/logger"
)

type Store struct {
	env              *env.Env
	logger           *logger.Logger
	events           *Events
	metricProcessors []interfaces.MetricProcesser
}

func (s Store) InitMetrics() {
	for _, metricProcessor := range s.metricProcessors {
		metricProcessor := metricProcessor
		metricProcessor.InitMetrics()
	}
}

func (s *Store) OnLogEntriesReceived(cluster string, server string, legacyFile *os.File, originalFile *os.File) {
	wg := &sync.WaitGroup{}
	for _, metricProcessor := range s.metricProcessors {
		metricProcessor := metricProcessor
		wg.Add(1)
		go func(metricProcessor interfaces.MetricProcesser, legacyFile *os.File, originalFile *os.File) {
			defer wg.Done()
			metricProcessor.ParseFile(legacyFile, originalFile, cluster, server)
		}(metricProcessor, legacyFile, originalFile)
	}
	wg.Wait()
	s.events.Publish(fmt.Sprintf(EVENT_LOGS_ENTRIES_STORED_BY_SERVER, cluster, server))
}

func (s *Store) OnAllServersReceived() {
	for _, metricProcessor := range s.metricProcessors {
		metricProcessor := metricProcessor
		metricProcessor.UpdateMetrics()
	}
}

func NewStore(
	env *env.Env,
	logger *logger.Logger,
	events *Events,
	logInfoMetric *metrics.LogInfoMetric,
	collectionScanMetric *metrics.CollectionScanMetric,
) *Store {
	s := &Store{
		env:    env,
		logger: logger,
		events: events,
		metricProcessors: []interfaces.MetricProcesser{
			logInfoMetric,
			collectionScanMetric,
		},
	}
	s.events.SubscribeAsync(EVENT_LOGS_SERVERS_RECEIVED, s.OnAllServersReceived, false)
	return s
}
