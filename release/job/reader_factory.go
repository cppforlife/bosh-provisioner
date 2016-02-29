package job

import (
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bpdload "github.com/cppforlife/bosh-provisioner/downloader"
	bptar "github.com/cppforlife/bosh-provisioner/tar"
)

const (
	readerFactoryDirPrefix = "dir://"
)

type ReaderFactory struct {
	downloader bpdload.Downloader
	extractor  bptar.Extractor
	fs         boshsys.FileSystem
	logger     boshlog.Logger
}

func NewReaderFactory(
	downloader bpdload.Downloader,
	extractor bptar.Extractor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ReaderFactory {
	return ReaderFactory{
		downloader: downloader,
		extractor:  extractor,
		fs:         fs,
		logger:     logger,
	}
}

func (rf ReaderFactory) NewReader(url string) Reader {
	if strings.HasPrefix(url, readerFactoryDirPrefix) {
		dir := url[len(readerFactoryDirPrefix):]
		return NewDirReader(dir, rf.fs, rf.logger)
	}

	return rf.NewTarReader(url)
}

func (rf ReaderFactory) NewTarReader(url string) Reader {
	return NewTarReader(url, rf.downloader, rf.extractor, rf.fs, rf.logger)
}
