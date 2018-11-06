package album

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"golang.org/x/net/html"
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

const nextPageID = "next_url"

// Page represents one page
type Page struct {
	nextPageURL string

	thumbnailSrc   string
	thumbnailID    string
	dynamicIndices [3]int

	pictureURL string
}

// New returns a new instance of Album
func New(logger *zap.SugaredLogger, rawURL string) (a Album, err error) {
	a.logger = logger
	a.url, err = url.Parse(rawURL)
	a.user = strings.Split(a.url.Path, "/")[1]
	a.logger.Infow("created album",
		"url", a.url, "user", a.user,
	)
	return
}

// Download downloads the album
func (a *Album) Download() (err error) {
	a.logger.Info("downloading the album", a)

	// fetch the source for the first page
	resp, err := http.Get(a.url.String())
	if err != nil {
		a.logger.DPanic("error fetching the source: ", err)
	}

	// parse the source
	page := Page{}
	tokenizer := html.NewTokenizer(resp.Body)
	endOfDocument := false
	for !endOfDocument {
		tokenType := tokenizer.Next()
		switch {
		case tokenType == html.ErrorToken:
			// end of document
			endOfDocument = true
			break
		case tokenType == html.TextToken:
			token := tokenizer.Token()
			// parse the title
			if strings.Contains(token.String(), "@iMGSRC.RU") {
				a.name = token.String()[:strings.LastIndex(token.String(), ",")]
				a.logger.Debug("parsed album name: ", a.name)
				continue
			}
			// parse the indices
			// i=q.slice(z+1,60)+e[7]+e.charAt(6)+e[2];
			const indiceRegex = `(?mU)^i=q\.slice\(z\+1,60\)\+.+(?:\[|\()(\d)(?:\]|\)).+(?:\[|\()(\d)(?:\]|\)).+(?:\[|\()(\d)(?:\]|\));$`
			iLiteral := "i=q.slice(z+1,60)+"
			if strings.Contains(token.String(), iLiteral) {
				r := regexp.MustCompile(indiceRegex)
				matches := r.FindAllStringSubmatch(token.String(), 3)
				a.logger.Debug(matches)
				// loop through the submatches and assign them
				for i := 0; i < 3; i++ {
					page.dynamicIndices[i], err = strconv.Atoi(matches[0][i+1])
					if err != nil {
						a.logger.Error(err)
					}
				}
			}
			break
		case tokenType == html.StartTagToken:
			token := tokenizer.Token()

			// check if current token is img or a tag
			if !((token.Data == "img") || (token.Data == "a")) {
				continue
			}

			// loop through all attributes
			for outerAttKey, outerAtt := range token.Attr {
				// get the thumbnail and the next page
				for _, innerAtt := range token.Attr {
					if (outerAtt.Key == "src" && innerAtt.Key == "id") || (outerAtt.Key == "id" && outerAtt.Val == nextPageID) {
						// print all the token
						a.logger.Debug(token)

						// next url
						if innerAtt.Key == "href" {
							page.nextPageURL = innerAtt.Val
						}

						// thumbnail
						if len(innerAtt.Val) == 8 && (innerAtt.Val != nextPageID) {
							page.thumbnailID = innerAtt.Val
							page.thumbnailSrc = token.Attr[outerAttKey].Val
						}
					}
				}
			}
		}
	}

	// do the magic
	err = page.FormatPictureLink()
	if err != nil {
		a.logger.DPanic(err)
		return
	}
	a.logger.Info(page)

	return
}

func ParsePage() {

}

// FormatPictureLink formats the picture link from all the attributes
func (p *Page) FormatPictureLink() (err error) {
	lastIndexOfSlash := strings.LastIndex(p.thumbnailSrc, "/")

	p.pictureURL = fmt.Sprintf(
		"https://b%s/%s%s%s%s.jpg",
		p.thumbnailSrc[9:lastIndexOfSlash],
		p.thumbnailSrc[lastIndexOfSlash+1:60],
		string(p.thumbnailID[p.dynamicIndices[0]]),
		string(p.thumbnailID[p.dynamicIndices[1]]),
		string(p.thumbnailID[p.dynamicIndices[2]]),
	)

	return
}

// GetFullName returns the full name of an album including the user
func (a *Album) GetFullName() string {
	return fmt.Sprintf("%s @ %s", a.user, a.name)
}
