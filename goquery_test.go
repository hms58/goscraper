// 2017/12/30 17:37:14 Sat
package goscraper

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hms58/dateparse"
	"github.com/hms58/goscraper"
)

func TestGoQueryFile(t *testing.T) {
	htmlFile := "test/jianshu.html"

	s, err := goscraper.Scrape(&goscraper.Options{
		Url:      "https://www.jianshu.com/p/fa24238d84a9",
		HtmlFile: htmlFile,
		Handler:  &goscraper.DefaultHandler{},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Printf("Title : %s\n", s.Preview.Title)
	// fmt.Printf("Description : %s\n", s.Preview.Description)
	// fmt.Printf("Image: %s\n", s.Preview.Images[0])
	fmt.Printf("Image: %d\n%v\n", len(s.Preview.Images), s.Preview.Images)
	fmt.Printf("RawImage: %d\n%v\n", len(s.Preview.RawImages), s.Preview.RawImages)
	fmt.Printf("RawScripts: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
	//
	fmt.Printf("title: %s\n", s.Preview.Title)

	fmt.Printf("body: %d\n", CountTextWords(s.Find("div[id=\"js_content\"]").Eq(0).Text()))
	fmt.Printf("body: %d\n", strings.Count(s.Find("div[id=\"js_content\"]").Eq(0).Text(), ""))
	publishTimeStr := s.Find("em[id=\"post-date\"]").Eq(0).Text()
	// publishTime, _ := time.Parse("2017-12-13", publishTimeStr)
	publishTime, _ := dateparse.ParseIn("2017-12-13", time.Local)
	// publishTime2, _ := dateparse.ParseAny("2017-12-13")
	fmt.Printf("publish_time: %s, %s, %s\n", publishTimeStr, publishTime.String())

	var topics []string
	s.Find("div[id=\"js_content\"]").Find("a").Each(func(index int, sel *goquery.Selection) {
		if val, ok := sel.Attr("href"); ok {
			fmt.Printf("title: %s, %s\n", sel.Text(), val)
			if isWechatTopic(val) {
				val = strings.Replace(val, "&", "&amp;", -1)
				topics = append(topics, val)
			}
		}
	})
	fmt.Printf("topics %d\n %v", len(topics), topics)

	var ahrefs []string
	s.Find("a").Each(func(index int, sel *goquery.Selection) {
		if val, ok := sel.Attr("href"); ok {
			// fmt.Printf("title: %s, %s\n", sel.Text(), val)
			if goscraper.IsImageUrl(val) {
				ahrefs = append(ahrefs, val)
			}
		}
	})
	fmt.Printf("a hrefs: %d\n%v\n", len(ahrefs), ahrefs)
}

func isWechatTopic(s string) bool {
	if strings.HasPrefix(s, "https://mp.weixin.qq.com/s?__biz=") ||
		strings.HasPrefix(s, "http://mp.weixin.qq.com/s?__biz=") {
		return true
	}
	return false
}

func TestGoqueryBody(t *testing.T) {
	body := `
		<html>
		<head>
		<link href="//mmbiz.qpic.cn" />
		<link href="https://mmbiz.qpic.cn/test/hello" />
		<link href="https://mmbiz.qpic.cn/test/good.css" />
		</head>
			<body>
				<img src="http://test.com/test.png?x=y"/>
				<img src="//test.com/test.png?x=y"/>
				<img src="/test.com/test.png?x=y"/>
				<img src="warning/test.png?x=y"/>
				<img src=""/>
				test
			</body>
		</html>
	`
	s, err := goscraper.Scrape(&goscraper.Options{
		Url:     "http://www.jianshu.com/p/99b7f266a7ec",
		Body:    body,
		Handler: &goscraper.DefaultHandler{},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Printf("Title : %s\n", s.Preview.Title)
	// fmt.Printf("Description : %s\n", s.Preview.Description)
	// fmt.Printf("Image: %s\n", s.Preview.Images[0])
	fmt.Printf("Image: %d\n%v\n", len(s.Preview.Images), s.Preview.Images)
	fmt.Printf("RawImage: %d\n%v\n", len(s.Preview.RawImages), s.Preview.RawImages)
	fmt.Printf("Js: %d\n%v\n", len(s.Preview.Scripts), s.Preview.Scripts)
	fmt.Printf("RawJs: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	text := s.Find("body").Eq(0).Text()

	fmt.Printf("body: %d\n", CountTextWords(text))
}

func CountTextWords(text string) int {
	skipCountLetter := "\r\n 	"
	count := 0
	for _, word := range text {
		// fmt.Printf("----%c----", word)
		if !strings.ContainsRune(skipCountLetter, word) {
			count = count + 1
		}
	}
	return count
}
