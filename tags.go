// 2017/12/29 23:05:54 Fri
package goscraper

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ValImage = iota
	ValScript
)

var imageExts = []string{
	".png",
	".jpg",
	".gif",
	".bmp",
	".tiff",
	".ico",
}

var scriptExts = []string{
	".js",
	".css",
}

var (
	resourceMap = make(map[string][]TagNode)
)

type TagHandler interface {
	// TagsFilter(string) ([]TagNode, bool)
	OnImage(*url.URL, string, string, string, *Selection) (string, bool)
	OnScript(*url.URL, string, string, string, *Selection) (string, bool)
}

type TagNode struct {
	ValType int
	Attrs   []string
}

func init() {

	// imgTag :=
	resourceMap["img"] = []TagNode{
		TagNode{
			ValType: ValImage,
			Attrs: []string{
				"src",
				"data-src",
				"data-original-src",
			},
		},
	}

	resourceMap["script"] = []TagNode{
		TagNode{
			ValType: ValScript,
			Attrs: []string{
				"src",
				"data-src",
			},
		},
	}

	resourceMap["link"] = []TagNode{
		TagNode{
			ValType: ValImage,
			Attrs: []string{
				"href",
			},
		},
		TagNode{
			ValType: ValScript,
			Attrs: []string{
				"href",
			},
		},
	}
}

func AddTagNode(tag string, node []TagNode) map[string][]TagNode {
	resourceMap[tag] = node
	return resourceMap
}

func (l TagNode) HasAttr(attr string) bool {
	for _, v := range l.Attrs {
		if v == attr {
			return true
		}
	}
	return false
}

type DefaultHandler struct {
	// tagsMap map[string][]TagNode
}

var _ TagHandler = (*DefaultHandler)(nil)

func (h *DefaultHandler) TagsFilter(tag string) ([]TagNode, bool) {
	nodes, ok := resourceMap[tag]
	return nodes, ok
}

func (h DefaultHandler) OnImage(parnUrl *url.URL, tag, key, val string, sel *Selection) (string, bool) {
	if tag == "link" {
		if !h.HasSuffix(val, imageExts) {
			return val, false
		}
	}
	return h.parseUrl(parnUrl, val)
}

func (h DefaultHandler) OnScript(parnUrl *url.URL, tag, key, val string, sel *Selection) (string, bool) {
	if tag == "link" {
		if !h.HasSuffix(val, scriptExts) {
			return val, false
		}
	}
	return h.parseUrl(parnUrl, val)
}

func (h DefaultHandler) parseUrl(parnUrl *url.URL, urlstr string) (string, bool) {

	u, err := url.Parse(urlstr)
	if err != nil {
		return urlstr, false
	}
	uri := u.RequestURI()
	var newUrl string
	if !u.IsAbs() {
		if u.Host == "" {
			newUrl = fmt.Sprintf("%s://%s%s", parnUrl.Scheme, parnUrl.Host, uri)
		} else {
			newUrl = fmt.Sprintf("%s://%s%s", parnUrl.Scheme, u.Host, uri)
		}
	} else {
		newUrl = urlstr
	}
	return newUrl, true
}

func (h DefaultHandler) HasSuffix(str string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(str, ext) {
			return true
		}
	}
	return false
}
