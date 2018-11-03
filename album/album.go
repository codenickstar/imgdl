package album

import (
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Album represents one image album
type Album struct {
	url *url.URL

	user         string
	name         string
	downloadPath string

	rawPictureURLs []string
	pictureURLs    []*url.URL

	logger *zap.SugaredLogger
}

// New returns a new instance of Album
func New(logger *zap.SugaredLogger, rawURL string) (a Album, err error) {
	a.logger = logger
	a.url, err = url.Parse(rawURL)
	a.user = strings.Split(a.url.Path, "/")[1]
	a.logger.Debugw("created new album",
		"url", a.url, "user", a.user,
	)
	return
}

// Download downloads the album
func (a *Album) Download() (err error) {
	// fetch the source
	// parse the source
	return
}

// GetFullName returns the full name of an album including the user
func (a *Album) GetFullName() string {
	return fmt.Sprintf("%s @ %s", a.user, a.name)
}
