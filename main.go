
package main
import (
	"net/http"
	"net/url"
	"fmt"
	"log"
	"strings"
	"io/ioutil"
)
import "github.com/PuerkitoBio/goquery"
import "github.com/gin-gonic/gin"
import "github.com/ryanuber/go-glob"

type APIResponse struct {
	Title   string `json:"title"`
	Indexed bool   `json:"indexed"`
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://tools.meettea.com")
		url := c.Query("url")
		if url == "" {
  		c.JSON(200, APIResponse{ "", false })
  		return
		}
		body := request(url)
		indexed := hasIndexed(url, body)
		c.JSON(200, indexed)
	})
	r.Run(":7002")
}

func hasIndexed(url string, body string) APIResponse {
	pos := strings.Index(body, "class=\"content_none\"")
	if pos  > -1 {
		return APIResponse { "", false }
	}
	pos = strings.Index(body, "没有找到该URL。您可以直接访问")
	if pos  > -1 {
		return  APIResponse { "", false }
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	firstBlock := doc.Find("#1")
	blockCount := firstBlock.Length()
	title := ""
	if blockCount > 0 {
		firstLinkText := firstBlock.Find(".c-showurl > b").Text()
		fmt.Println("firstLinkText", firstLinkText)
		if firstLinkText == "" {
			return  APIResponse { "", false }
		}
		urlWithoutProtocal := removeUrlProtocol(url)
		title = firstBlock.Find("h3").Text()
		if strings.HasSuffix(firstLinkText, "/") {
			firstLinkText = firstLinkText[ 0 : len(firstLinkText) - 1 ]
		}
		if strings.Index(firstLinkText, "...") > -1 {
			firstLinkText = strings.Replace(firstLinkText, "...", "*", -1)
		}
		firstLinkText += "*"
		indexed := glob.Glob(firstLinkText, urlWithoutProtocal)
		return APIResponse { title, indexed }
	} else {
		log.Fatalln("未找到 div#1 块, body: ", body)
	}
	return APIResponse { title, true }
}

func removeUrlProtocol(url string) string {
	protocol := "://"
	pos := strings.Index(url, protocol)
	if pos == -1 {
		return url
	}
	return url[pos + len(protocol):]
}

func request(queryUrl string) string {
	resp, err := http.Get("http://www.baidu.com/s?wd=" + url.QueryEscape(queryUrl))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}