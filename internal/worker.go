package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	hashCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	hashLength  = 84
)

type Item struct {
	Linkcode string `json:"linkcode"`
	Type     int    `json:"type"` // 0 = folder, 1 = file
}

type Link struct {
	Next string `json:"next"`
}

type Response struct {
	Items []Item `json:"items"`
	Link  Link   `json:"_links"`
}

type RequestPayload struct {
	Link   string `json:"link"`
	Cookie string `json:"cookie"`
}

var DOMAIN = "https://www.fshare.vn"
var DOMAIN_API string = fmt.Sprintf("%s/api", DOMAIN)
var DOMAIN_FILE string = fmt.Sprintf("%s/file", DOMAIN)

func init() {
	// Initialize random seed for better randomness
	rand.Seed(time.Now().UnixNano())
}

// User agents pool to rotate and avoid detection
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// getRandomUserAgent returns a random user agent from the pool
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// getRandomDelay returns a random delay between 500ms and 2000ms to simulate human behavior
func getRandomDelay() time.Duration {
	delay := rand.Intn(1500) + 500 // 500-2000ms
	return time.Duration(delay) * time.Millisecond
}

func getLink(folderId string, page int, perPage int) Response {
	// Add random delay to avoid rate limiting
	time.Sleep(getRandomDelay())

	url := fmt.Sprintf("%s/v3/files/folder?linkcode=%s&sort=type,name&page=%d&per-page=%d", DOMAIN_API, folderId, page, perPage)
	
	// Create HTTP client with timeout and proper redirect handling
	client := http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects but preserve headers
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	// Set realistic browser headers to avoid bot detection
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("https://www.fshare.vn/folder/%s", folderId))
	req.Header.Set("Origin", "https://www.fshare.vn")
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	res, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer res.Body.Close()

	// Check for rate limiting or blocking
	if res.StatusCode == 429 {
		// Rate limited, wait longer before retrying
		fmt.Printf("Rate limited, waiting 5 seconds...\n")
		time.Sleep(5 * time.Second)
		// Retry once
		return getLink(folderId, page, perPage)
	}

	if res.StatusCode != 200 {
		panic(fmt.Sprintf("Unexpected status code: %d", res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	var resp Response
	json.Unmarshal(body, &resp)
	return resp
}

func checkFilOrFolder(url string) bool {
	return strings.Contains(url, "/folder/")
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

	link, err := FetchFshareLinks(
		fmt.Sprintf("%s&pass=undefined&hash=%s&captcha=undefined", payload.Link, randomHash),
		payload.Cookie,
	)
	if err != nil {
		return "", err
	}
	return link, nil
}

// ProcessLinksWorker processes links and returns VIP links
func ProcessLinksWorker(links []string, cookie string, getFolderIdFromUrl func(string) string) []string {
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

	// Recursive function to process folders and handle nested folders
	var processFolder func(folderId string, page int, perPage int)
	processFolder = func(folderId string, page int, perPage int) {
		resp := getLink(folderId, page, perPage)
		if len(resp.Items) == 0 {
			return
		}

		for _, item := range resp.Items {
			counter += 1
			if item.Type == 0 {
				// Type 0 = folder, recursively process it
				fmt.Printf("Found folder with linkcode: %s, processing recursively...\n", item.Linkcode)
				// Add delay before processing nested folder to avoid rapid requests
				time.Sleep(getRandomDelay())
				processFolder(item.Linkcode, 1, perPage)
			} else if item.Type == 1 {
				// Type 1 = file, add to fileUrls
				fileUrls = append(fileUrls, fmt.Sprintf("%s/%s", DOMAIN_FILE, item.Linkcode))
			}
		}

		// Handle pagination - if there's a next page, process it
		if resp.Link.Next != "" {
			// Add delay before processing next page
			time.Sleep(getRandomDelay())
			processFolder(folderId, page+1, perPage)
		}
	}

	// Process all initial folders
	for i, folderId := range folderIds {
		if i > 0 {
			// Add delay between processing different folders
			time.Sleep(getRandomDelay())
		}
		processFolder(folderId, page, perPage)
	}
	fileUrls = uniqArray(fileUrls)
	linkVips := []string{}
	for _, fileUrl := range fileUrls {
		// sleep 3-7 seconds
		sleep := rand.Intn(100) + 3000                      // 3-7 seconds
		time.Sleep(time.Duration(sleep) * time.Millisecond) // sleep 3-7 seconds
		linkVip, err := linkVipDotNet(fileUrl, cookie)
		fmt.Printf("linkVip: %s\n", linkVip)
		if err != nil {
			fmt.Printf("link %s error: %s\n", fileUrl, err.Error())
			continue
		}
		fmt.Printf("%s \n", fileUrl)
		linkVips = append(linkVips, linkVip)
	}
	return uniqArray(linkVips)
}

