package main

import (
	"flag"
	"fmt"
	"net/http"
	"./videos"
	"./actresses"
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

	switch mode {
	case "videos":
		videos.Parse(req.Body)
	case "actresses":
		actresses.Parse(req.Body)
	}
}
