package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

type Item struct {
	Linkcode string `json:"linkcode"`
}

type Link struct {
	Next string `json:"next"`
}

type Response struct {
	Items []Item `json:"items"`
	Link  Link   `json:"_links"`
}

type File struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

var DOMAIN = "https://www.fshare.vn"
var DOMAIN_API string = fmt.Sprintf("%s/api", DOMAIN)
var DOMAIN_FILE string = fmt.Sprintf("%s/file", DOMAIN)

func getLink(folderId string, page int, perPage int) Response {
	url := fmt.Sprintf("%s/v3/files/folder?linkcode=%s&sort=type,name&page=%d&per-page=%d", DOMAIN_API, folderId, page, perPage)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"Content-Type":    {"application/json"},
		"accept":          {"application/json, text/plain, */*"},
		"accept-language": {"vi-VN,vi"},
		"referer":         {fmt.Sprintf("https://www.fshare.vn/folder/%s", folderId)},
		"user-agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"},
	}
	fmt.Println(url)
	if err != nil {
		panic(err.Error())
	}
	res, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	var resp Response
	json.Unmarshal(body, &resp)
	return resp
}

func getLinks(c echo.Context) error {
	folders := c.QueryParam("folders")
	folderIds := strings.Split(folders, ",")
	page := 1
	perPage := 50
	fileUrls := []string{}
	counter := 0
	var recursion func(folderIds []string, page int, perPage int)
	recursion = func(folderIds []string, page int, perPage int) {
		for _, folderId := range folderIds {
			firstCallResp := getLink(folderId, page, perPage)
			if len(firstCallResp.Items) == 0 {
				break
			} else {
				for _, item := range firstCallResp.Items {
					counter += 1
					fileUrls = append(fileUrls, fmt.Sprintf("%s/%s", DOMAIN_FILE, item.Linkcode))
				}
				if firstCallResp.Link.Next != "" {
					recursion(folderIds, page+1, perPage)
				}
			}
		}
	}
	recursion(folderIds, page, perPage)
	return c.JSONPretty(http.StatusOK, fileUrls, "")
}

func main() {
	// http://localhost:9090/get-links?folders=54X131MQJ887,1GY93EHX1ZLQ
	e := echo.New()
	e.GET("/get-links", getLinks)
	e.Logger.Fatal(e.Start(":9090"))
}
