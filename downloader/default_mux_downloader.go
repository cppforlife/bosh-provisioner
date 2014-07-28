package downloader

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func NewDefaultMuxDownloader(
	blobstore boshblob.Blobstore,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) MuxDownloader {
	mux := map[string]Downloader{
		"http":      NewHTTPDownloader(fs, logger),
		"https":     NewHTTPDownloader(fs, logger),
		"file":      NewLocalFSDownloader(fs, logger),
		"blobstore": NewBlobstoreDownloader(blobstore, logger),
	}

	return NewMuxDownloader(mux, logger)
}
