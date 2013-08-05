package feedfind

import (
	"code.google.com/p/go.net/html"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var FeedMIMETypes = map[string]bool{
	"application/x.atom+xml": true,
	"application/atom+xml":   true,
	"application/xml":        true,
	"text/xml":               true,
	"application/rss+xml":    true,
	"application/rdf+xml":    true,
}

func isValidFeedMIME(mimeType string) bool {
	_, ok := FeedMIMETypes[mimeType]
	return ok
}

var feedExtention = regexp.MustCompile(`\.(?:rss|xml|rdf)$`)

func isFeedUrl(url string) bool {
	return feedExtention.MatchString(url)
}

var spacesRegexp = regexp.MustCompile(`\s+`)

func findAttr(node *html.Node, attrName string) (string, bool) {
	for _, a := range node.Attr {
		if a.Key == attrName {
			return a.Val, true
		}
	}

	return "", false
}

func handleLinkTag(node *html.Node) (string, bool) {
	relAttr, ok := findAttr(node, "rel")
	if !ok {
		return "", false
	}

	rels := make(map[string]bool)
	for _, rel := range spacesRegexp.Split(relAttr, -1) {
		rels[rel] = true
	}

	typeAttr, ok := findAttr(node, "type")
	if !ok {
		return "", false
	}

	hrefAttr, ok := findAttr(node, "href")
	if !ok {
		return "", false
	}

	typeAttr = strings.Trim(typeAttr, "")
	_, relok := rels["alternate"]
	_, srvok := rels["service.feed"]
	if isValidFeedMIME(typeAttr) && (relok || srvok) {
		return hrefAttr, true
	}

	return "", false
}

func toAbsoluteUrl(base *url.URL, urlStr string) string {
	relUrl, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	absUrl := base.ResolveReference(relUrl)
	return absUrl.String()
}

var ignoreTagRegexp = regexp.MustCompile(`^(?:meta|isindex|title|script|style|head|html)$`)

func Find(base string) ([]string, error) {
	res, err := http.Get(base)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	baseUrl, err := url.Parse(base)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	feeds := make([]string, 0)
	endFlag := false

	var nodeWalk func(*html.Node)
	nodeWalk = func(n *html.Node) {
		if endFlag {
			return
		}

		if n.Type == html.ElementNode {
			tagName := n.Data

			if tagName == "link" {
				feed, ok := handleLinkTag(n)
				if !ok {
					return
				}

				feedAbsUrl := toAbsoluteUrl(baseUrl, feed)
				feeds = append(feeds, feedAbsUrl)
			} else if tagName == "base" {
				href, ok := findAttr(n, "href")
				if !ok {
					return
				}

				baseUrl, err = url.Parse(href)
				if err != nil {
					return
				}
			} else if ignoreTagRegexp.MatchString(tagName) {
				// Ignore other valid tags inside of <head>
			} else if tagName == "a" {
				href, ok := findAttr(n, "href")
				if !ok {
					return
				}

				feedAbsUrl := toAbsoluteUrl(baseUrl, href)
				if isFeedUrl(feedAbsUrl) {
					feeds = append(feeds, feedAbsUrl)
				}
			} else {
				if len(feeds) > 0 {
					endFlag = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			nodeWalk(c)
		}
	}
	nodeWalk(doc)

	return feeds, nil
}
