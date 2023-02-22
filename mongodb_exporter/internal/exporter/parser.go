package exporter

import (
	"mongodb_performance_exporter/interfaces"
	"mongodb_performance_exporter/legacy"
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

	// create a temporary file to store the logs
	originalFile, fErr := os.CreateTemp("tmp/logs", "mongo-logs-")
	if fErr != nil {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": fErr}).Error("Unable to create temporary file for logs")
	}
	originalFile.WriteString(logs)

	// convert to legacy format
	legacyFile, fErr := os.CreateTemp("tmp/logs", "mongo-legacy-logs-")
	if fErr != nil {
		p.logger.WithFields(logrus.Fields{"cluster": cluster, "server": server, "error": fErr}).Error("Unable to create temporary file for legacy logs")
	}

	legacyFilePath := legacyFile.Name()
	converter := &legacy.LogConverter{}
	converter.ParseFile(originalFile.Name(), &legacyFilePath)

	//
	// close and remove the temporary file at the end of the program
	defer legacyFile.Close()
	defer originalFile.Close()
	defer os.Remove(legacyFile.Name())
	defer os.Remove(originalFile.Name())

	p.store.OnLogEntriesReceived(cluster, server, legacyFile)
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
