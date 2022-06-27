package exporter

import (
	go_app "github.com/shoplineapp/go-app"
	"github.com/shoplineapp/go-app/plugins/logger"
)

type ExporterModule struct {
	go_app.AppModuleInterface

	Crawler *Crawler
	Metric  *Metric
}

func (m *ExporterModule) Controllers() []interface{} {
	return []interface{}{}
}

func (m *ExporterModule) Provide() []interface{} {
	return []interface{}{
		func(
			logger *logger.Logger,
			parser *Parser,
			crawler *Crawler,
			store *Store,
			metric *Metric,
		) *ExporterModule {
			return &ExporterModule{
				Crawler: crawler,
				Metric:  metric,
			}
		},
		NewParser,
		NewCrawler,
		NewStore,
		NewMetric,
		NewEventBus,
	}
}
