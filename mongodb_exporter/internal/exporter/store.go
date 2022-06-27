package exporter

import (
	"fmt"

	"github.com/shoplineapp/go-app/plugins/env"
	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type Store struct {
	env    *env.Env
	logger *logger.Logger
	events *Events

	LogInfoQueries []LogInfoQuery
	buffer         []LogInfoQuery
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

func (e LogInfoQuery) Labels() []string {
	return []string{
		e.Cluster,
		e.Server,
		e.Namespace,
		e.Operation,
		e.Pattern,
	}
}

func (s *Store) OnLogEntriesReceived(cluster string, server string, entries []LogInfoQuery) {
	s.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(entries)}).Debug("Log entries saved into buffer")
	s.buffer = append(s.buffer, entries...)
	s.events.Publish(fmt.Sprintf(EVENT_LOGS_ENTRIES_STORED_BY_SERVER, cluster, server))
}

func (s *Store) OnAllServersReceived() {
	s.logger.WithFields(logrus.Fields{"buffer": len(s.buffer)}).Debug("Flushing buffer to store")
	for _, entry := range s.LogInfoQueries {
		MongoAtlasTopQueryCount.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMinMS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMaxMS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryP95MS.DeleteLabelValues(entry.Labels()...)
		MongoAtlasTopQueryMeanMS.DeleteLabelValues(entry.Labels()...)
	}

	s.LogInfoQueries = s.buffer

	s.logger.WithFields(logrus.Fields{"count": len(s.LogInfoQueries)}).Info("Reporting data to metric API")
	for _, entry := range s.LogInfoQueries {
		MongoAtlasTopQueryCount.WithLabelValues(entry.Labels()...).Set(float64(entry.Count))
		MongoAtlasTopQueryMinMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MinMS))
		MongoAtlasTopQueryMaxMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MaxMS))
		MongoAtlasTopQueryP95MS.WithLabelValues(entry.Labels()...).Set(float64(entry.P95MS))
		MongoAtlasTopQueryMeanMS.WithLabelValues(entry.Labels()...).Set(float64(entry.MeanMS))
	}
	s.buffer = []LogInfoQuery{}
}

func NewStore(env *env.Env, logger *logger.Logger, events *Events) *Store {
	store := &Store{
		env:            env,
		logger:         logger,
		events:         events,
		LogInfoQueries: []LogInfoQuery{},
		buffer:         []LogInfoQuery{},
	}
	store.events.SubscribeAsync(EVENT_LOGS_ENTRIES_PARSED, store.OnLogEntriesReceived, false)
	store.events.SubscribeAsync(EVENT_LOGS_SERVERS_RECEIVED, store.OnAllServersReceived, false)
	return store
}
