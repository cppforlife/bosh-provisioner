package downloader

import (
	"io"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const httpDownloaderLogTag = "HTTPDownloader"

type HTTPDownloader struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewHTTPDownloader(
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) HTTPDownloader {
	return HTTPDownloader{fs: fs, logger: logger}
}

func (d HTTPDownloader) Download(url string) (string, error) {
	file, err := d.fs.TempFile("release-Downloader-HTTPDownloader")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating download destination")
	}

	d.logger.Debug(httpDownloaderLogTag, "Downloaded %s to %s", url, file.Name())

	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", bosherr.WrapError(err, "Get url")
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", bosherr.WrapError(err, "Copying response to file")
	}

	return file.Name(), nil
}

func (d HTTPDownloader) CleanUp(path string) error {
	return d.fs.RemoveAll(path)
}
