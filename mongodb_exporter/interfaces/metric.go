package interfaces

import "os"

type MetricProcesser interface {
	InitMetrics()
	UpdateMetrics()
	ParseFile(file *os.File, cluster string, server string)
}
