package interfaces

import "os"

type MetricProcesser interface {
	InitMetrics()
	UpdateMetrics()
	ParseFile(legacyFile *os.File, originalFile *os.File, cluster string, server string)
}
