package exporter

import (
	"mongodb_performance_exporter/interfaces"
	"os"

	"github.com/shoplineapp/go-app/plugins/logger"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	logger *logger.Logger
	events *Events
	store  *Store
}

func (p *Parser) ParseLogs(cluster string, server string, logs string) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "err": err}).Error("Failed to parse logs")
			p.events.Publish(EVENT_LOGS_ENTRIES_PARSED, cluster, server, []interfaces.MetricProcesser{})
		}
	}()

	p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(logs)}).Debug("Parsing logs")
	if len(logs) == 0 {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "count": len(logs)}).Debug("No log to be processed, skipping")
		p.events.Publish(EVENT_LOGS_ENTRIES_PARSED, cluster, server, []interfaces.MetricProcesser{})
		return
	}

	f, fErr := os.CreateTemp("tmp/logs", "mongo-logs-")
	if fErr != nil {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": fErr}).Error("Unable to create temporary file for logs")
	}
	f.WriteString(logs)

	// close and remove the temporary file at the end of the program
	defer f.Close()
	defer os.Remove(f.Name())

	p.store.OnLogEntriesReceived(cluster, server, f)
}

func NewParser(logger *logger.Logger, events *Events, store *Store) *Parser {
	p := &Parser{
		logger: logger,
		events: events,
		store:  store,
	}
	p.events.SubscribeAsync(EVENT_LOGS_DOWNLOADED, p.ParseLogs, false)
	return p
}
