package album

import (
	"fmt"
	"net/http"
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

	pagesCount int
	pages      []Page

	logger *zap.SugaredLogger
}

// New returns a new instance of Album
func New(logger *zap.SugaredLogger, rawURL string) (a Album, err error) {
	a.logger = logger
	a.url, err = url.Parse(rawURL)
	a.user = strings.Split(a.url.Path, "/")[1]
	a.logger.Infow("created album", "url", a.url, "user", a.user)
	return
}

// Download downloads the album
func (a *Album) Download() (err error) {
	a.logger.Info("downloading the album: ", a)

	// first url
	nextURL := a.url.String()

	// parse the source
	for {
		// fetch the source
		resp, err := http.Get(nextURL)
		if err != nil {
			a.logger.DPanic("error fetching the source: ", err)
			return err
		}

		page, err := ParsePage(a, resp.Body)
		if err != nil {
			a.logger.DPanic("error parsing page: ", err)
			return err
		}

		a.logger.Debug("adding page: ", page)
		a.pages = append(a.pages, page)

		nextURL = page.Next()
		a.logger.Debug("next url: ", nextURL)
		if nextURL == "" {
			break
		}
	}

	for _, page := range a.pages {
		fmt.Println(page.pictureURL)
	}

	return
}

// GetFullName returns the full name of an album including the user
func (a *Album) GetFullName() string {
	return fmt.Sprintf("%s @ %s", a.user, a.name)
}
