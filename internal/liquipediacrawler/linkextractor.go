package liquipediacrawler

import (
	"bytes"
	"cmp"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/yusufaine/cs3203-g46-crawler/internal/crawler"
)

var URLRegex = regexp.MustCompile(`[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,24}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

// Returns a list of all outgoing links from the page
func ReportLinkExtractor(c *crawler.Crawler, currURL *url.URL, resp []byte) []*url.URL {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		log.Error("unable to parse response body", "error", err)
		return nil
	}

	linkSet := map[string]*url.URL{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if !ok {
			return
		}

		// skip if link cannot be parsed
		newUrl, err := url.Parse(link)
		if err != nil {
			return
		}

		if !strings.Contains(newUrl.Path, "dota2/The_International") {
			return
		}

		updatedURL := *currURL
		updatedURL.Path = newUrl.Path

		linkSet[updatedURL.String()] = &updatedURL
	})

	urls := make([]*url.URL, 0, len(linkSet))
	for _, v := range linkSet {
		urls = append(urls, v)
	}
	slices.SortFunc(urls, func(a, b *url.URL) int {
		return cmp.Compare(a.String(), b.String())
	})

	return urls
}

// Returns a list of all outgoing links that have not been visited from the page
func TIAnalyserLinkExtractor(c *crawler.Crawler, currURL *url.URL, resp []byte) []*url.URL {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		log.Error("unable to parse response body", "error", err)
		return nil
	}

	linkSet := map[string]*url.URL{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if !ok {
			return
		}

		// skip if link cannot be parsed
		newUrl, err := url.Parse(link)
		if err != nil {
			return
		}

		if !strings.Contains(newUrl.Path, "dota2/The_International") {
			return
		}

		updatedURL := *currURL
		updatedURL.Path = newUrl.Path

		c.PageMutex.Lock()
		defer c.PageMutex.Unlock()
		if _, ok := c.VisitedPageInfo[updatedURL.String()]; ok {
			return
		}

		linkSet[updatedURL.String()] = &updatedURL
	})

	urls := make([]*url.URL, 0, len(linkSet))
	for _, v := range linkSet {
		urls = append(urls, v)
	}
	slices.SortFunc(urls, func(a, b *url.URL) int {
		return cmp.Compare(a.String(), b.String())
	})

	return urls
}