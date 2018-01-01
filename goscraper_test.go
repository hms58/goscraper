// 2017/12/29 20:44:28 Fri
package goscraper

import (
	"fmt"
	"testing"

	"github.com/hms58/goscraper"
)

// var imgRE = regexp.MustCompile(`<img[^>]+\bsrc=["']([^"']+)["']`)
// // if your img's are properly formed with doublequotes then use this, it's more efficient.
// // var imgRE = regexp.MustCompile(`<img[^>]+\bsrc="([^"]+)"`)
// func findImages(htm string) []string {
//     imgs := imgRE.FindAllStringSubmatch(htm, -1)
//     out := make([]string, len(imgs))
//     for i := range out {
//         out[i] = imgs[i][1]
//     }
//     return out
// }

// func main() {
func testGoscraper(t *testing.T) {
	s, err := goscraper.ScrapeRedirect(&goscraper.Options{
		Url:     "https://www.jianshu.com/p/569bdd440b09",
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
	fmt.Printf("RawScripts: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
}

func TestGoscraperBody(t *testing.T) {
	body := `
		<html>
			<body>
				<img src="http://test.com/test.png?x=y"/>
				<img src="//test.com/test.png?x=y"/>
				<img src="/test.com/test.png?x=y"/>
				<img src="warning/test.png?x=y"/>
			</body>
		</html>
	`
	s, err := goscraper.ScrapeRedirect(&goscraper.Options{
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
	fmt.Printf("Image: %d\n%v\n", len(s.Preview.RawImages), s.Preview.RawImages)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
}

func TestGoscraperFile(t *testing.T) {
	htmlFile := "test/jianshu.html"

	s, err := goscraper.ScrapeRedirect(&goscraper.Options{
		Url:      "https://www.jianshu.com/p/fa24238d84a9",
		HtmlFile: htmlFile,
		Handler:  &goscraper.DefaultHandler{},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Title : %s\n", s.Preview.Title)
	// fmt.Printf("Description : %s\n", s.Preview.Description)
	// fmt.Printf("Image: %s\n", s.Preview.Images[0])
	// fmt.Printf("Image: %d\n%v\n", len(s.Preview.Images), s.Preview.Images)
	// fmt.Printf("RawImage: %d\n%v\n", len(s.Preview.RawImages), s.Preview.RawImages)
	fmt.Printf("RawScripts: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
}

func TestGoscraperFile2(t *testing.T) {
	htmlFile := "test/wechat.html"

	s, err := goscraper.ScrapeRedirect(&goscraper.Options{
		Url:      "https://mp.weixin.qq.com/s?__biz=MjM5ODYxMDA5OQ==&mid=2651960726&idx=1&sn=0fdaf0e7040318aabfeba553f815d691&chksm=bd2d004a8a5a895ca80180443cc0f18e66b3d15dbbbd120dabaf3e6d4ef00fbc1030bf41c24b&scene=21",
		HtmlFile: htmlFile,
		Handler:  &goscraper.DefaultHandler{},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Title : %s\n", s.Preview.Title)
	// fmt.Printf("Description : %s\n", s.Preview.Description)
	// fmt.Printf("Image: %s\n", s.Preview.Images[0])
	// fmt.Printf("Image: %d\n%v\n", len(s.Preview.Images), s.Preview.Images)
	fmt.Printf("RawImage: %d\n%v\n", len(s.Preview.RawImages), s.Preview.RawImages)
	fmt.Printf("RawScripts: %d\n%v\n", len(s.Preview.RawScripts), s.Preview.RawScripts)
	// fmt.Printf("Url : %s\n", s.Preview.Link)
}
