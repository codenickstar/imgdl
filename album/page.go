package album

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// the constants
const nextPageID = "next_url"
const indiceRegex = `(?mU)^i=q\.slice\(z\+1,60\)\+.+(?:\[|\()(\d)(?:\]|\)).+(?:\[|\()(\d)(?:\]|\)).+(?:\[|\()(\d)(?:\]|\));$`

// Page represents one page
type Page struct {
	nextPageURL string

	thumbnailSrc   string
	thumbnailID    string
	dynamicIndices [3]int

	pictureURL string
}

// Next returns the next page url to parse
func (p *Page) Next() string {
	return p.nextPageURL
}

// ParsePage parses the given body and returns the filled page struct
func ParsePage(album *Album, body io.Reader) (page Page, err error) {
	tokenizer := html.NewTokenizer(body)
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
			if album.name == "" {
				if strings.Contains(token.String(), "@iMGSRC.RU") {
					album.name = token.String()[:strings.LastIndex(token.String(), ",")]
					album.logger.Debug("parsed album name: ", album.name)
					continue
				}
			}

			// parse the indices
			// i=q.slice(z+1,60)+e[7]+e.charAt(6)+e[2];
			if strings.Contains(token.String(), "i=q.slice(z+1,60)+") {
				r := regexp.MustCompile(indiceRegex)
				matches := r.FindAllStringSubmatch(token.String(), 3)
				// album.logger.Debug(matches)
				// loop through the submatches and assign them
				for i := 0; i < 3; i++ {
					page.dynamicIndices[i], err = strconv.Atoi(matches[0][i+1])
					if err != nil {
						album.logger.Error(err)
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
						// album.logger.Debug(token)

						// next url
						if innerAtt.Key == "href" {
							if !strings.HasPrefix(innerAtt.Val, "/main/user.php?user=") {
								page.nextPageURL = fmt.Sprintf("http://imgsrc.ru%s", innerAtt.Val)
							}
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
	err = page.formatPictureLink()
	if err != nil {
		album.logger.DPanic(err)
		return
	}
	return
}

// formatPictureLink formats the picture link from all the attributes
func (p *Page) formatPictureLink() (err error) {
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
