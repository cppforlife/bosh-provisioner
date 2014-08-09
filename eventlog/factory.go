package eventlog

import (
	"os"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Factory struct {
	config Config
	logger boshlog.Logger
}

func NewFactory(config Config, logger boshlog.Logger) Factory {
	return Factory{config: config, logger: logger}
}

func (f Factory) NewLog() Log {
	var device Device

	switch f.config.DeviceType {
	case ConfigDeviceTypeJSON:
		device = NewJSONDevice(os.Stdout)
	case ConfigDeviceTypeText:
		device = NewTextDevice(os.Stdout)
	default:
		// config should be validated before using it with a factory
		panic(bosherr.New("Unknown device type '%s'", f.config.DeviceType))
	}

	return NewLog(device, f.logger)
}
