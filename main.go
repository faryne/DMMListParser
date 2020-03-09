package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type DmmVideosListError struct {
	Error 	string 				`json:"error"`
	ErrorNo string 				`json:"status"`
}

type DmmVideosList struct {
	Videos 	[]DmmVideo	`json:"videos"`
}

type DmmVideo struct {
	DmmVideoHeader
	DmmVideoBody
}

type DmmVideoHeader struct {
	No 		string 	`json:"no"`
	Title 	string	`json:"title"`
	Url 	string 	`json:"url"`
	Thumb 	string	`json:"thumb"`
}

type DmmVideoBody struct {
	VodDate			string 	`json:"vod_date"`
	PublishDate		string 	`json:"pulish_date"`
	Duration 		int 	`json:"duration"`
	Directors 		[]string 	`json:"directors"`
	Series			[]string 	`json:"series"`
	Makers 			[]string 	`json:"makers"`
	Labels 			[]string 	`json:"labels"`
	Tags 			[]string 	`json:"tags"`
	Actresses 		[]string 	`json:"actresses"`
	Images 			[]DmmVideoImage	`json:"images"`
}

type DmmVideoImage struct {
	Thumb 			string	`json:"thumb"`
	Preview 		string	`json:"preview"`
}

func main () {
	var url string
	flag.StringVar(&url,"dmmurl", "", "dmm 列表頁網址")
	flag.Parse()

	// 判斷網址參數
	if len(url) == 0 {
		panic("網址不得為空")
	}



	// 建立 http 連線以便取得網頁內容
	req, err := http.Get(url)
	// 如果發生錯誤時
	if err != nil {
		fmt.Printf(">>> Error: %s\n", err)
	}
	//
	// 不是 200 時
	if req.StatusCode != 200 {
		fmt.Printf(">>> Error: %d\n", req.StatusCode)
	}
	defer req.Body.Close()
	docs, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		fmt.Printf(">>> Error: %s\n", err)
	}
	imgs := docs.Find("p.tmb > a")


	// 輸出內容定在這裡
	videos := []DmmVideo{}
	var ch = make(chan DmmVideoBody)

	imgs.Each(func(idx int, s *goquery.Selection){
		// 解出網址 / 縮圖網址 / 標題
		link, _ := s.Attr("href")
		thumb, _ := s.Find("span > img").Attr("src")
		title, _ := s.Find("span > img").Attr("alt")

		// 解出番號
		NoPattern, _ := regexp.Compile("cid=([^/]+)")
		NoString := NoPattern.FindStringSubmatch(link)

		VideoHeader := DmmVideoHeader{
			No: strings.ToUpper(NoString[1]),
			Url: link,
			Title: title,
			Thumb: thumb,
		}

		go func () { ch <- ParsePage(link) }()

		videos = append(videos, DmmVideo{
			VideoHeader,
			<-ch,
		})
	})

	output := DmmVideosList{
		Videos: videos,
	}
	j, _ := json.Marshal(output)
	fmt.Printf("%s\n", j)
}

/**
 * 解析網頁內容
 */
func ParsePage (PageUrl string) DmmVideoBody {
	// 定義 pattern
	Patterns := map[string]*regexp.Regexp{}
	Patterns["vod_date"], _ = regexp.Compile("配信開始")
	Patterns["publish_date"], _ = regexp.Compile("商品発売")
	Patterns["duration"], _ = regexp.Compile("収録時間")
	Patterns["directors"], _ = regexp.Compile("監督")
	Patterns["series"], _ = regexp.Compile("シリーズ")
	Patterns["makers"], _ = regexp.Compile("メーカー")
	Patterns["labels"], _ = regexp.Compile("レーベル")
	Patterns["tags"], _ = regexp.Compile("ジャンル")

	var VideoBody = DmmVideoBody{}
	// 解出詳細資訊
	req1, _ := http.Get(PageUrl)
	defer req1.Body.Close()

	v, _ := goquery.NewDocumentFromReader(req1.Body)

	VideoBody.Actresses = ParseActresses(*v, PageUrl)
	VideoBody.Images = parseImages(*v)

	rows := v.Find("table.mg-b20 > tbody > tr")
	rows.Each(func(idx int, e *goquery.Selection){
		rowTitle, _ := e.Find("td.nw").Html()
		rowValue := e.Find("td:last-child")

		for k, v := range Patterns {
			if v.MatchString(rowTitle) {
				rowContent, _ := rowValue.Html()
				switch k {
				case "vod_date":
					VideoBody.VodDate = strings.TrimSpace(rowContent)
					break
				case "publish_date":
					VideoBody.PublishDate = strings.TrimSpace(rowContent)
					break
				case "duration":
					DurationPattern, _ := regexp.Compile("([0-9]+)")
					duration := DurationPattern.FindStringSubmatch(rowContent)
					VideoBody.Duration, _ = strconv.Atoi(strings.TrimSpace(duration[1]))
					break
				case "directors":
					VideoBody.Directors = make([]string, 0)
					rowValue.Find("a").Each(func(i int, d *goquery.Selection) {
						h, _ := d.Html()
						VideoBody.Directors =  append(VideoBody.Directors, strings.TrimSpace(h))
					})
					break
				case "series":
					VideoBody.Series = make([]string, 0)
					rowValue.Find("a").Each(func(i int, d *goquery.Selection) {
						h, _ := d.Html()
						VideoBody.Series =  append(VideoBody.Series, strings.TrimSpace(h))
					})
					break
				case "makers":
					VideoBody.Makers = make([]string, 0)
					rowValue.Find("a").Each(func(i int, d *goquery.Selection) {
						h, _ := d.Html()
						VideoBody.Makers =  append(VideoBody.Makers, strings.TrimSpace(h))
					})
				case "labels":
					VideoBody.Labels = make([]string, 0)
					rowValue.Find("a").Each(func(i int, d *goquery.Selection) {
						h, _ := d.Html()
						VideoBody.Labels =  append(VideoBody.Labels, strings.TrimSpace(h))
					})
				case "tags":
					VideoBody.Tags = make([]string, 0)
					rowValue.Find("a").Each(func(i int, d *goquery.Selection) {
						h, _ := d.Html()
						VideoBody.Tags = append(VideoBody.Tags, strings.TrimSpace(h))
					})
				default:
					fmt.Printf("----\n")
				}
			}
		}

	})

	return VideoBody
}

/**
 * 解析女優列表資訊
 */
func ParseActresses (document goquery.Document, url string) []string {
	element := document.Find("a#a_performer")

	var actresses = make([]string, 0)
	if element.Length() == 0 {
		document.Find("span#performer > a").Each(func(i int, s *goquery.Selection){
			h, _ := s.Html()
			actresses = append(actresses, strings.TrimSpace(h))
		})
	} else {
		// 在網頁中找看看有沒有這個網址 pattern
		reg, _ := regexp.Compile(`'(/digital/videoa/-/detail/ajax-performer[^']+)'`)
		html, _ := document.Html()
		matches := reg.FindStringSubmatch(html)

		// 取得標籤列表網頁
		TagUrl := "https://www.dmm.co.jp" + matches[1]


		Client := http.Client{}
		req, _ := http.NewRequest("GET", TagUrl, strings.NewReader(""))
		req.Header.Add("Referer", url)
		resp, _ :=Client.Do(req)
		defer resp.Body.Close()

		parser, _ := goquery.NewDocumentFromReader(resp.Body)

		parser.Find("a").Each(func(idx int, s *goquery.Selection){
			h, _ := s.Html()
			actresses = append(actresses, strings.TrimSpace(h))
		})
	}
	return actresses
}

/**
 * 解析網頁中的預覽圖片
 */
func parseImages (document goquery.Document) []DmmVideoImage {
	images := document.Find("img.mg-b6")

	var OututImages = []DmmVideoImage{}

	images.Each(func(idx int, s *goquery.Selection){
		thumb, _ := s.Attr("src")

		pattern, _ := regexp.Compile("(\\-[0-9]+\\.jpg)$")
		preview := pattern.ReplaceAllString(thumb, `jp${1}`)

		imageBody := DmmVideoImage{
			Preview: preview,
			Thumb: thumb,
		}

		OututImages = append(OututImages, imageBody)
	})

	return OututImages
}
