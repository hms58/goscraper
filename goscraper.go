package goscraper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	EscapedFragment string = "_escaped_fragment_="
	fragmentRegexp         = regexp.MustCompile("#!(.*)")
)

type Options struct {
	Url         string
	MaxRedirect int
	Body        string
	Handler     TagHandler
	HtmlFile    string
}

type Scraper struct {
	Url                *url.URL
	EscapedFragmentUrl *url.URL
	Body               []byte
	MaxRedirect        int
	handler            TagHandler
}

type DocumentPreview struct {
	Title       string
	Description string
	Images      []string
	RawImages   []string
	Scripts     []string
	RawScripts  []string

	Link string
}

// func Scrape(uri, body string, maxRedirect int, handler TagHandler) (*Document, error) {
func ScrapeRedirect(opts *Options) (*Document, error) {
	u, err := url.Parse(opts.Url)
	if err != nil {
		return nil, err
	}
	var body []byte
	if opts.HtmlFile != "" {
		content, err := ioutil.ReadFile(opts.HtmlFile)
		if err != nil {
			return nil, err
		}
		// _, name, _ := charset.DetermineEncoding(content, "")
		// l.Info("type: %s", name)
		body = content
	} else {
		body = []byte(opts.Body[:])
	}
	// if len(body) == 0 {
	// 	body = []byte(opts.Body[:])
	// }
	return (&Scraper{
		Url:         u,
		Body:        body,
		MaxRedirect: opts.MaxRedirect,
		handler:     opts.Handler,
	}).Scrape()
}

func (scraper *Scraper) Scrape() (*Document, error) {
	var doc *Document
	var err error

	if len(scraper.Body) == 0 {
		doc, err = scraper.getDocument()
	} else {
		doc, err = scraper.getDocumentFromBody()
	}
	if err != nil {
		return nil, err
	}
	err = scraper.parseDocument(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (scraper *Scraper) getUrl() string {
	if scraper.EscapedFragmentUrl != nil {
		return scraper.EscapedFragmentUrl.String()
	}
	return scraper.Url.String()
}

func (scraper *Scraper) toFragmentUrl() error {
	unescapedurl, err := url.QueryUnescape(scraper.Url.String())
	if err != nil {
		return err
	}
	matches := fragmentRegexp.FindStringSubmatch(unescapedurl)
	if len(matches) > 1 {
		escapedFragment := EscapedFragment
		for _, r := range matches[1] {
			b := byte(r)
			if avoidByte(b) {
				continue
			}
			if escapeByte(b) {
				escapedFragment += url.QueryEscape(string(r))
			} else {
				escapedFragment += string(r)
			}
		}

		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(strings.Replace(unescapedurl, matches[0], p+escapedFragment, 1))
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	} else {
		p := "?"
		if len(scraper.Url.Query()) > 0 {
			p = "&"
		}
		fragmentUrl, err := url.Parse(unescapedurl + p + EscapedFragment)
		if err != nil {
			return err
		}
		scraper.EscapedFragmentUrl = fragmentUrl
	}
	return nil
}

func (scraper *Scraper) getDocument() (*Document, error) {
	scraper.MaxRedirect -= 1
	if strings.Contains(scraper.Url.String(), "#!") {
		scraper.toFragmentUrl()
	}
	if strings.Contains(scraper.Url.String(), EscapedFragment) {
		scraper.EscapedFragmentUrl = scraper.Url
	}

	req, err := http.NewRequest("GET", scraper.getUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "GoScraper")

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if resp.Request.URL.String() != scraper.getUrl() {
		scraper.EscapedFragmentUrl = nil
		scraper.Url = resp.Request.URL
	}
	b, err := convertUTF8(resp.Body)
	if err != nil {
		return nil, err
	}
	doc := &Document{Body: b, Preview: DocumentPreview{Link: scraper.Url.String()}}

	return doc, nil
}

func (scraper *Scraper) getDocumentFromBody() (*Document, error) {
	b := bytes.NewBuffer(scraper.Body)

	// b, err := convertUTF8(bytes.NewReader(scraper.Body))
	// if err != nil {
	// 	return nil, err
	// }

	doc := &Document{Body: b, Preview: DocumentPreview{Link: scraper.Url.String()}}
	return doc, nil
}

func convertUTF8(content io.Reader) (*bytes.Buffer, error) {
	buff := &bytes.Buffer{}
	content, err := charset.NewReader(content, "")
	if err != nil {
		return buff, err
	}
	_, err = io.Copy(buff, content)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func (scraper *Scraper) parseDocument(doc *Document) error {
	t := html.NewTokenizer(doc.Body)
	var ogImage bool
	var headPassed bool
	var hasFragment bool
	var hasCanonical bool
	var canonicalUrl *url.URL
	doc.Preview.Images = []string{}
	// saves previews' link in case that <link rel="canonical"> is found after <meta property="og:url">
	link := doc.Preview.Link
	for {
		tokenType := t.Next()
		if tokenType == html.ErrorToken {
			return nil
		}
		if tokenType != html.SelfClosingTagToken && tokenType != html.StartTagToken && tokenType != html.EndTagToken {
			continue
		}
		token := t.Token()

		switch token.Data {
		case "head":
			if tokenType == html.EndTagToken {
				headPassed = true
			}
		case "body":
			headPassed = true

		case "link":
			if scraper.handler != nil {
				break
			}
			var canonical bool
			var href string
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "rel" && cleanStr(attr.Val) == "canonical" {
					canonical = true
				}
				if cleanStr(attr.Key) == "href" {
					href = attr.Val
				}
				if len(href) > 0 && canonical && link != href {
					hasCanonical = true
					var err error
					canonicalUrl, err = url.Parse(href)
					if err != nil {
						return err
					}
				}
			}

		case "meta":
			if len(token.Attr) != 2 {
				break
			}
			if metaFragment(token) && scraper.EscapedFragmentUrl == nil {
				hasFragment = true
			}
			var property string
			var content string
			for _, attr := range token.Attr {
				if cleanStr(attr.Key) == "property" || cleanStr(attr.Key) == "name" {
					property = attr.Val
				} else if cleanStr(attr.Key) == "content" {
					content = attr.Val
				}
			}
			switch cleanStr(property) {
			case "og:title":
				doc.Preview.Title = content
			case "og:description":
				doc.Preview.Description = content
			case "description":
				if len(doc.Preview.Description) == 0 {
					doc.Preview.Description = content
				}
			case "og:url":
				doc.Preview.Link = content
			case "og:image":
				ogImage = true
				if scraper.handler != nil {
					attrMap := make(map[string]string)
					newUrl, ok := scraper.handler.OnImage(scraper.Url, token.Data, "og:image", content, &Selection{AttrMap: attrMap})
					if ok {
						doc.Preview.Images = append(doc.Preview.Images, newUrl)
						doc.Preview.RawImages = append(doc.Preview.RawImages, content)
					}
				} else {
					ogImgUrl, err := url.Parse(content)
					if err != nil {
						return err
					}
					if !ogImgUrl.IsAbs() {
						ogImgUrl, err = url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, ogImgUrl.Path))
						if err != nil {
							return err
						}
					}

					doc.Preview.Images = []string{ogImgUrl.String()}
				}

			}

		case "title":
			if tokenType == html.StartTagToken {
				t.Next()
				token = t.Token()
				if len(doc.Preview.Title) == 0 {
					doc.Preview.Title = token.Data
				}
			}

		case "img":
			// l.Info("%v", token.Attr)
			if scraper.handler != nil {
				break
			}
			for _, attr := range token.Attr {
				key := cleanStr(attr.Key)
				val := strings.TrimSpace(attr.Val)
				if key == "" || val == "" {
					continue
				}

				if key == "src" {
					imgUrl, err := url.Parse(attr.Val)
					if err != nil {
						return err
					}
					uri := imgUrl.RequestURI()
					// l.Debug("%s", uri)
					if !imgUrl.IsAbs() {
						if imgUrl.Host == "" {
							doc.Preview.Images = append(doc.Preview.Images, fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, uri))
						} else {
							doc.Preview.Images = append(doc.Preview.Images, fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, imgUrl.Host, uri))
						}

					} else {
						doc.Preview.Images = append(doc.Preview.Images, attr.Val)
					}
				}
			}
		}

		if scraper.handler != nil {
			if nodes, ok := resourceMap[token.Data]; ok {
				// if nodes, ok := scraper.handler.TagsFilter(token.Data); ok {
				attrMap := make(map[string]string, len(token.Attr))
				for _, attr := range token.Attr {
					key := cleanStr(attr.Key)
					attrMap[key] = attr.Val
				}
				for _, node := range nodes {
					for _, key := range node.Attrs {
						if val, ok := attrMap[key]; ok {
							nval := strings.TrimSpace(val)
							switch node.ValType {
							case ValImage:
								newUrl, ok := scraper.handler.OnImage(scraper.Url, token.Data, key, nval, &Selection{AttrMap: attrMap})
								if ok {
									doc.Preview.Images = append(doc.Preview.Images, newUrl)
									doc.Preview.RawImages = append(doc.Preview.RawImages, val)
								}
							case ValScript:
								newUrl, ok := scraper.handler.OnScript(scraper.Url, token.Data, key, nval, &Selection{AttrMap: attrMap})
								if ok {
									doc.Preview.Scripts = append(doc.Preview.Scripts, newUrl)
									doc.Preview.RawScripts = append(doc.Preview.RawScripts, val)
								}
							}
						}
					}
				}
				// for _, attr := range token.Attr {
				// 	key := cleanStr(attr.Key)
				// 	val := strings.TrimSpace(attr.Val)
				// 	for _, v := range nodes {
				// 		if v.HasAttr(key) {
				// 			switch v.ValType {
				// 			case ValImage:
				// 				newUrl, _, ok := scraper.handler.OnImage(scraper.Url, token.Data, key, val)
				// 				if ok {
				// 					doc.Preview.Images = append(doc.Preview.Images, newUrl)
				// 					doc.Preview.RawImages = append(doc.Preview.RawImages, attr.Val)
				// 				}
				// 			case ValScript:
				// 				newUrl, _, ok := scraper.handler.OnScript(scraper.Url, token.Data, key, val)
				// 				if ok {
				// 					doc.Preview.Scripts = append(doc.Preview.Scripts, newUrl)
				// 					doc.Preview.RawScripts = append(doc.Preview.RawScripts, attr.Val)
				// 				}
				// 			}
				// 		}
				// 	}
				// }
			}
		}
		if hasCanonical && headPassed && scraper.MaxRedirect > 0 {
			if !canonicalUrl.IsAbs() {
				absCanonical, err := url.Parse(fmt.Sprintf("%s://%s%s", scraper.Url.Scheme, scraper.Url.Host, canonicalUrl.Path))
				if err != nil {
					return err
				}
				canonicalUrl = absCanonical
			}
			scraper.Url = canonicalUrl
			scraper.EscapedFragmentUrl = nil
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if hasFragment && headPassed && scraper.MaxRedirect > 0 {
			scraper.toFragmentUrl()
			fdoc, err := scraper.getDocument()
			if err != nil {
				return err
			}
			*doc = *fdoc
			return scraper.parseDocument(doc)
		}

		if len(doc.Preview.Title) > 0 && len(doc.Preview.Description) > 0 && ogImage && headPassed {
			return nil
		}

	}

	return nil
}

func avoidByte(b byte) bool {
	i := int(b)
	if i == 127 || (i >= 0 && i <= 31) {
		return true
	}
	return false
}

func escapeByte(b byte) bool {
	i := int(b)
	if i == 32 || i == 35 || i == 37 || i == 38 || i == 43 || (i >= 127 && i <= 255) {
		return true
	}
	return false
}

func metaFragment(token html.Token) bool {
	var name string
	var content string

	for _, attr := range token.Attr {
		if cleanStr(attr.Key) == "name" {
			name = attr.Val
		}
		if cleanStr(attr.Key) == "content" {
			content = attr.Val
		}
	}
	if name == "fragment" && content == "!" {
		return true
	}
	return false
}

func cleanStr(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}
