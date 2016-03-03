package updater

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bpagclient "github.com/cppforlife/bosh-provisioner/agent/client"
)

const postStarterLogTag = "PostStarter"

var (
	ErrPostScriptFailed = bosherr.Error("Post start scripts failed")
)

type PostStarter struct {
	agentClient bpagclient.Client
	logger      boshlog.Logger
}

func NewPostStarter(
	agentClient bpagclient.Client,
	logger boshlog.Logger,
) PostStarter {
	return PostStarter{
		agentClient: agentClient,
		logger:      logger,
	}
}

// PostStart runs after an instance reaches running state.
func (w PostStarter) PostStart() error {
	w.logger.Debug(starterLogTag, "Running post-start")

	_, err := w.agentClient.PostStart()
	if err != nil {
		return ErrPostScriptFailed
	}

	return nil
}
