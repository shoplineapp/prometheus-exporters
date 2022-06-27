package main

import (
	"mongodb_performance_exporter/internal/exporter"

	go_app "github.com/shoplineapp/go-app"
	"github.com/shoplineapp/go-app/plugins/env"
)

func main() {
	app := go_app.NewApplication()

	app.AddModule(&exporter.ExporterModule{})

	app.Run(func(
		env *env.Env,
		exporter *exporter.ExporterModule,
	) {
		env.SetDefaultEnv(map[string]string{
			"CRAWLER_INTERVAL_TIME": "1h",
			"CRAWLER_SINCE_TIME":    "1h",
		})
		exporter.Crawler.Listen()
		exporter.Metric.Serve()
	})
}
