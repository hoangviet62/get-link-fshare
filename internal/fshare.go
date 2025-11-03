package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
)

type Message struct {
	Text string `json:"text"`
}

// Trangthai handles both string and int values from JSON
type Trangthai string

func (t *Trangthai) UnmarshalJSON(data []byte) error {
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*t = Trangthai(fmt.Sprintf("%d", intVal))
		return nil
	}
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}
	*t = Trangthai(strVal)
	return nil
}

type APIResponse struct {
	Trangthai Trangthai `json:"trangthai"`
	Filename  string    `json:"filename"`
	Linkvip   string    `json:"linkvip"`
	Loi       string    `json:"loi"`
}

type FinalResponse struct {
	Messages []Message `json:"messages"`
}

func FetchFshareLinks(link string, cookie string) (string, error) {
	url := "https://linksvip.net/GetLinkFs"
	result, err := fileGetContentsCurl(url, 5, link, cookie)
	if err != nil {
		return "", err
	}

	cleanedResult := cleanString(result)

	var apiResp APIResponse
	if err := json.Unmarshal([]byte(cleanedResult), &apiResp); err != nil {
		fmt.Println(cleanedResult)
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	var finalResp FinalResponse

	switch apiResp.Trangthai {
	case "1":
		finalResp = FinalResponse{
			Messages: []Message{
				{Text: "Link của bạn đã sẵn sàng để download. <3"},
				{Text: "*Tên file:* " + apiResp.Filename},
				{Text: "*Download:* " + apiResp.Linkvip},
			},
		}
	default:
		finalResp = FinalResponse{
			Messages: []Message{
				{Text: apiResp.Loi},
			},
		}
	}

	return finalResp.Messages[0].Text, nil
}

func cleanString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\r' || r == '\n' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func fileGetContentsCurl(url string, retries int, post string, cookie string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	var req *http.Request
	var err error

	if post != "" {
		post = fmt.Sprintf("link=%s", post)
		req, err = http.NewRequest("POST", url, strings.NewReader(post))
		if err != nil {
			return "", fmt.Errorf("failed to create POST request: %w", err)
		}
	} else {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create GET request: %w", err)
		}
	}

	// Set headers (equivalent to CURLOPT_HTTPHEADER)
	req.Header.Set("Host", "linksvip.net")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://linksvip.net")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("DNT", "1")
	req.Header.Set("Referer", "https://linksvip.net/")
	req.Header.Set("Cookie", "user=vietnth0602%40gmail.com; pass=78321e89c3e254e911a18c4de61837b1; __stripe_mid=8a67f84a-cd6d-4e2e-9e01-71423add8e89ee2a3e; PHPSESSID=t428nskme2v45bnrn74esv2he2; _csrf=N08sIThAMkUxIS5YOEoyQDFTLiM3IzdSMUQuWDNTMUcxVCxW; __stripe_sid=d234cbc6-907a-45ef-9130-069957214938e4452c")

	resp, err := client.Do(req)

	if err != nil {
		if retries > 0 {
			time.Sleep(1 * time.Second)
			return fileGetContentsCurl(url, retries-1, post, cookie)
		}
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if retries > 0 {
			time.Sleep(1 * time.Second)
			return fileGetContentsCurl(url, retries-1, post, cookie)
		}
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	result := string(body)

	if result == "" && retries > 0 {
		time.Sleep(1 * time.Second)
		return fileGetContentsCurl(url, retries-1, post, cookie)
	}

	return result, nil
}
