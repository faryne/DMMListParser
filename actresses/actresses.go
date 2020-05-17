package actresses

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"

	"io"
	"regexp"
	"strconv"
	"strings"
)

type DMMActress struct {
	Name 			string		`json:"name",default:""`
	Kana 			string		`json:"kana",default:""`
	Photo 			string		`json:"photo"`
	Height			int			`json:"height"`
	Bust 			int 		`json:"bust"`
	Waist 			int 		`json:"waist"`
	Hips 			int 		`json:"hips"`
	Cup 			string 		`json:"cup"`
	Horoscope		string 		`json:"horoscope"`
	Blood 			string 		`json:"blood"`
	BornCity 		string 		`json:"born_city"`
	BirthYear 		int 		`json:"birth_year"`
	BirthMonth 		int			`json:"birth_month"`
	BirthDay 		int 		`json:"birth_day"`
	FullBirthday 	string 		`json:"full_birthday"`
	Interests 		[]string 	`json:"interests"`
}

func Parse (reader io.Reader) {
	// https://github.com/djimenez/iconv-go 轉換出來的內容可能有問題，改用 charset 試試
	utfBody, e := charset.NewReader(reader, "text/html")
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	if utfBody == nil {
		fmt.Println("xxx")
		return
	}
	docs, err := goquery.NewDocumentFromReader(utfBody)

	// 1054998 / 9920  會有問題
	//fmt.Println(docs.Html())
	if err != nil {
		fmt.Printf(">>> Error: %s\n", err.Error())
	}

	// 初始化輸出物件
	actress := &DMMActress{}


	// 姓名
	name_field, e := docs.Find("td.t1 > h1").Html()
	if e != nil {
		fmt.Println("Error when parse name: ", e.Error())
	}
	// 為了避免正規式難懂，把全形括號換成底線
	replacer := strings.NewReplacer(
		"（", "_",
		"）", "_")
	name_field = replacer.Replace(name_field)
	pattern_name, _ := regexp.Compile("(?P<Name>[^（]*)_(?P<Kana>[^_]*)_")
	if pattern_name.MatchString(name_field) {
		match := pattern_name.FindStringSubmatch(name_field)
		for i, names := range pattern_name.SubexpNames() {
			switch (names) {
			case "Name":
				actress.Name = match[i]
			case "Kana":
				actress.Kana = match[i]
			}
		}
	}
	// 大頭照
	actress.Photo, _ = docs.Find("tr.area-av30.top > td:nth-child(1) > img").Attr("src")

	// 其他個人資料
	pattern_horoscope, _ := regexp.Compile(`星座`)
	pattern_blood, _ := regexp.Compile(`血液型`)
	pattern_born, _ := regexp.Compile(`出身地`)
	pattern_interests, _ := regexp.Compile(`趣味・特技`)
	pattern_body, _ := regexp.Compile(`サイズ`)
	pattern_day, _ := regexp.Compile(`生年月日`)
	pattern_3size, _ := regexp.Compile(`(T(?P<Height>[0-9]+)cm\s)?(B(?P<Bust>[0-9]+)cm)?(\((?P<Cup>[A-Z]{1,})カップ\))?(\sW(?P<Waist>[0-9]+)cm)?(\sH(?P<Hips>[0-9]+)cm)?`)
	pattern_birthday, _ := regexp.Compile(`((?P<Year>[0-9]{4})年)?((?P<Month>[0-9]{1,})月)?((?P<Day>[0-9]{1,})日)?`)

	docs.Find("tr.area-av30.top > td:nth-child(2) > table > tbody > tr").Each(func(i int, s *goquery.Selection){
		header, _ := s.Find("td:nth-child(1)").Html()
		value, _ := s.Find("td:nth-child(2)").Html()
		//fmt.Println(header)

		if pattern_horoscope.MatchString(header) {
			if value != "----" {
				actress.Horoscope = value
			}
		} else if pattern_blood.MatchString(header) {
			if value != "----" {
				actress.Blood = value
			}
		} else if pattern_born.MatchString(header) {
			if value != "----" {
				actress.BornCity = value
			}
		} else if pattern_interests.MatchString(header) {
			actress.Interests = make([]string, 0)
			interests := strings.Split(value, "、")

			if len(interests) > 0 {
				for _, data := range interests {
					if data != "----" {
						actress.Interests = append(actress.Interests, data)
					}
				}
			}
		} else if pattern_body.MatchString(header) {
			match_3size := pattern_3size.FindStringSubmatch(value)
			for i, names := range pattern_3size.SubexpNames() {
				switch (names) {
				case "Height":
					actress.Height, _ = strconv.Atoi(match_3size[i])
				case "Bust":
					actress.Bust, e = strconv.Atoi(match_3size[i])
				case "Cup":
					actress.Cup = match_3size[i]
				case "Waist":
					actress.Waist, _ = strconv.Atoi(match_3size[i])
				case "Hips":
					actress.Hips, _ = strconv.Atoi(match_3size[i])
				}
			}
		} else if pattern_day.MatchString(header) {
			match_bday := pattern_birthday.FindStringSubmatch(value)
			for i, names := range pattern_birthday.SubexpNames() {
				switch (names) {
				case "Year":
					actress.BirthYear, _ = strconv.Atoi(match_bday[i])
				case "Month":
					actress.BirthMonth, _ = strconv.Atoi(match_bday[i])
				case "Day":
					actress.BirthDay, _ = strconv.Atoi(match_bday[i])
				}
			}
			if actress.BirthYear > 0 && actress.BirthMonth > 0 && actress.BirthDay > 0 {
				actress.FullBirthday = fmt.Sprintf("%04d/%02d/%02d", actress.BirthYear, actress.BirthMonth, actress.BirthDay)
			}
		}

		//fmt.Println("Field: ", header)
		//fmt.Println("Value: ", value)
	})

	//fmt.Println(replacer.Replace(name))
	j, _ := json.Marshal(actress)
	fmt.Println(string(j))
}
