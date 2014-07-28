package updater

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bpagclient "github.com/cppforlife/bosh-provisioner/agent/client"
)

const stopperLogTag = "Stopper"

type Stopper struct {
	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewStopper(
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) Stopper {
	return Stopper{
		agentClient: agentClient,
		logger:      logger,
	}
}

func (s Stopper) Stop() error {
	s.logger.Debug(stopperLogTag, "Stopping instance")

	_, err := s.agentClient.Stop()
	if err != nil {
		return bosherr.WrapError(err, "Stopping")
	}

	return nil
}
