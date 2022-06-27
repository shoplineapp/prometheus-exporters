package exporter

import (
	evbus "github.com/asaskevich/EventBus"
)

const (
	EVENT_LOGS_DOWNLOADED               = "logs_downloaded"
	EVENT_LOGS_ENTRIES_PARSED           = "logs_entries_parsed"
	EVENT_LOGS_ENTRIES_STORED_BY_SERVER = "logs_entries_stored_by_server_%s_%s"
	EVENT_LOGS_SERVERS_RECEIVED         = "logs_servers_received"
)

type Events struct {
	evbus.Bus
}

func NewEventBus() *Events {
	return &Events{
		Bus: evbus.New(),
	}
}
