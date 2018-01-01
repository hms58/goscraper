// 2017/12/30 17:37:14 Sat
package goscraper

import (
	"fmt"
	"testing"

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
	// fmt.Printf("RawScripts: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
	//
	fmt.Printf("title: %s\n", s.Preview.Title)

	fmt.Printf("body: %s\n", s.Find("body").Eq(0).Text()[:20])
}
