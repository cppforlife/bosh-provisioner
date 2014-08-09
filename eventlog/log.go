package eventlog

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const logLogTag = "Log"

type Log struct {
	device Device
	logger boshlog.Logger
}

type LogEntry struct {
	Time int64 `json:"time"`

	Stage string   `json:"stage"`
	Task  string   `json:"task"`
	Tags  []string `json:"tags"`

	Total int `json:"total"`
	Index int `json:"index"`

	State    string `json:"state"`
	Progress int    `json:"progress"`

	// Might contain error key
	Data map[string]interface{} `json:"data,omitempty"`
}

func NewLog(device Device, logger boshlog.Logger) Log {
	return Log{device: device, logger: logger}
}

func (l Log) BeginStage(name string, total int) *Stage {
	return &Stage{
		log:   l,
		name:  name,
		total: total,
	}
}

func (l Log) WriteLogEntryNoErr(entry LogEntry) {
	err := l.device.WriteLogEntry(entry)
	if err != nil {
		l.logger.Error(logLogTag, "Failed writing log entry %s", err)
	}
}
