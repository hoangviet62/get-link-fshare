package main

import (
	"encoding/json"
	"fmt"
	"get-link-fshare/internal"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

const (
	hashCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	hashLength  = 84
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

type GetLinksRequest struct {
	Links  []string `json:"links"`
	Cookie string   `json:"cookie"`
}

// RequestPayload structure for incoming JSON payload
type RequestPayload struct {
	Link   string `json:"link"`
	Cookie string `json:"cookie"`
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
	var req GetLinksRequest
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read body"})
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
	}
	links := req.Links
	cookie := req.Cookie

	fileUrls := []string{}
	folderIds := []string{}
	for _, link := range links {
		if checkFilOrFolder(link) {
			folderId := getFolderIdFromUrl(link)
			folderIds = append(folderIds, folderId)
		} else {
			fileUrls = append(fileUrls, link)
		}
	}
	page := 1
	perPage := 50
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
	fileUrls = uniqArray(fileUrls)
	linkVips := []string{}
	for _, fileUrl := range fileUrls {
		// sleep 5-10 seconds
		sleep := rand.Intn(100) + 5000                      // 5-10 seconds
		time.Sleep(time.Duration(sleep) * time.Millisecond) // sleep 5-10 seconds
		linkVip, err := linkVipDotNet(fileUrl, cookie)
		if err != nil {
			fmt.Printf("link %s error: %s\n", fileUrl, err.Error())
			continue
		}
		fmt.Printf("%s \n", fileUrl)
		linkVips = append(linkVips, linkVip)
	}
	return c.JSONPretty(http.StatusOK, uniqArray(linkVips), "")
}

func checkFilOrFolder(url string) bool {
	return strings.Contains(url, "/folder/")
}

func getFolderIdFromUrl(url string) string {
	parts := strings.Split(url, "/folder/")
	return parts[1]
}

func uniqArray(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func generateRandomHash() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, hashLength)
	for i := range b {
		b[i] = hashCharset[rand.Intn(len(hashCharset))]
	}
	return string(b)
}

func linkVipDotNet(fileUrl string, cookie string) (string, error) {
	// Parse JSON payload
	var payload RequestPayload
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"link": "%s", "cookie": "%s"}`, fileUrl, cookie)), &payload); err != nil {
		return "", fmt.Errorf("failed to parse JSON payload: %w", err)
	}

	// Validate required fields
	if payload.Link == "" {
		return "", fmt.Errorf("link is required")
	}

	if payload.Cookie == "" {
		return "", fmt.Errorf("cookie is required")
	}

	randomHash := generateRandomHash()

	link, err := internal.FetchFshareLinks(
		fmt.Sprintf("%s&pass=undefined&hash=%s&captcha=undefined", payload.Link, randomHash),
		payload.Cookie,
	)
	if err != nil {
		return "", err
	}
	return link, nil
}

func main() {
	e := echo.New()
	e.POST("/get-links", getLinks)
	e.Logger.Fatal(e.Start(":8080"))
}
