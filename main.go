package main

import (
	"./actresses"
	"./videos"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type DmmVideosListError struct {
	Error 	string 				`json:"error"`
	ErrorNo string 				`json:"status"`
}

func main () {
	var url, mode string
	flag.StringVar(&url,"dmmurl", "", "dmm 列表頁網址")
	flag.StringVar(&mode, "mode", "videos", "解析模式")
	flag.Parse()

	// 判斷網址參數
	if len(url) == 0 {
		panic("網址不得為空")
	}

	// 建立 http 連線以便取得網頁內容
	c := http.Client{}
	var resp io.Reader
	req, _ := http.NewRequest("GET", url, resp)
	req.AddCookie(&http.Cookie{
		Name:       "age_check_done",
		Value:      "1",
	})
	out, err2 := c.Do(req)
	// 如果發生錯誤時
	if err2 != nil {
		fmt.Printf(">>> Error: %s\n", err2)
		return
	}
	//
	// 不是 200 時
	if out.StatusCode != 200 {
		fmt.Printf(">>> Error (HTTP): %d\n", out.StatusCode)
		responseContent, _ := ioutil.ReadAll(out.Body)
		fmt.Printf(string(responseContent))
		return
	}

	defer out.Body.Close()

	switch mode {
	case "videos":
		videos.Parse(out.Body)
	case "actresses":
		actresses.Parse(out.Body)
	}
}
