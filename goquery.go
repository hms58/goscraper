// 2017/12/30 17:11:13 Sat
package goscraper

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type HtmlParser struct {
	Url                *url.URL
	EscapedFragmentUrl *url.URL
	Body               []byte
	MaxRedirect        int
	handler            TagHandler
	Preview            DocumentPreview
	TagsMap            map[string][]TagNode
}

type Selection struct {
	Sel     *goquery.Selection
	AttrMap map[string]string
}

type Document struct {
	*goquery.Document
	Body    *bytes.Buffer
	Preview DocumentPreview
}

func (s Selection) Attr(name string) (string, bool) {
	if s.Sel != nil {
		return s.Sel.Attr(name)
	} else if s.AttrMap != nil {
		val, ok := s.AttrMap[name]
		return val, ok
	}
	return "", false
}

func Scrape(opts *Options) (*Document, error) {
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
		body = content
	} else {
		body = []byte(opts.Body[:])
	}
	if opts.Handler == nil {
		opts.Handler = &DefaultHandler{}
	}
	return (&HtmlParser{
		Url:         u,
		Body:        body,
		MaxRedirect: opts.MaxRedirect,
		handler:     opts.Handler,
		TagsMap:     resourceMap,
	}).Scrape()
}

func NewScrape(opts *Options) (*HtmlParser, error) {
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
		body = content
	} else {
		body = []byte(opts.Body[:])
	}
	if opts.Handler == nil {
		opts.Handler = &DefaultHandler{}
	}
	return &HtmlParser{
		Url:         u,
		Body:        body,
		MaxRedirect: opts.MaxRedirect,
		handler:     opts.Handler,
		TagsMap:     resourceMap,
	}, nil
}

func (scraper *HtmlParser) Scrape() (*Document, error) {
	doc, err := scraper.getDocument()
	if err != nil {
		return nil, err
	}
	err = scraper.parseDocument(doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (scraper *HtmlParser) AddTagNode(tag string, node []TagNode) {
	scraper.TagsMap[tag] = node
	return
}

func (scraper *HtmlParser) getDocument() (*Document, error) {
	if len(scraper.Body) == 0 {
		doc, err := goquery.NewDocument(scraper.getUrl())
		if err != nil {
			return nil, err
		}
		return &Document{Document: doc, Preview: DocumentPreview{Link: scraper.Url.String()}}, nil
	} else {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(scraper.Body))
		if err != nil {
			return nil, err
		}
		return &Document{Document: doc, Preview: DocumentPreview{Link: scraper.Url.String()}}, nil
	}
	panic("Unexcept")
}

func (scraper *HtmlParser) getUrl() string {
	return scraper.Url.String()
}

func (scraper *HtmlParser) parseDocument(doc *Document) error {

	if scraper.handler == nil {
		return nil
	}

	doc.Preview.Title = doc.Find("title").Eq(0).Text()

	for tagName, tagNodes := range scraper.TagsMap {

		doc.Find(tagName).Each(func(index int, sel *goquery.Selection) {
			for _, node := range tagNodes {
				for _, attr := range node.Attrs {
					if val, ok := sel.Attr(attr); ok {
						switch node.ValType {
						case ValImage:
							newUrl, ok := scraper.handler.OnImage(scraper.Url, tagName, attr, val, &Selection{Sel: sel})
							if ok {
								doc.Preview.Images = append(doc.Preview.Images, newUrl)
								doc.Preview.RawImages = append(doc.Preview.RawImages, val)
							}
						case ValScript:
							newUrl, ok := scraper.handler.OnScript(scraper.Url, tagName, attr, val, &Selection{Sel: sel})
							if ok {
								doc.Preview.Scripts = append(doc.Preview.Scripts, newUrl)
								doc.Preview.RawScripts = append(doc.Preview.RawScripts, val)
							}
						}

						// l.Info("%s = %s", attr, val)
					}
				}
			}
		})
	}
	return nil
}
