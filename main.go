package main

import (
	"encoding/json"
	"get-link-fshare/internal"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo"
)

type GetLinksRequest struct {
	Links  []string `json:"links"`
	Cookie string   `json:"cookie"`
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

	// Use worker to process links
	linkVips := internal.ProcessLinksWorker(req.Links, req.Cookie, getFolderIdFromUrl)
	return c.JSONPretty(http.StatusOK, linkVips, "")
}

func getFolderIdFromUrl(urlStr string) string {
	// Parse the URL to handle query parameters
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// Fallback to simple string splitting if URL parsing fails
		parts := strings.Split(urlStr, "/folder/")
		if len(parts) < 2 {
			return ""
		}
		// Remove query parameters manually
		folderId := strings.Split(parts[1], "?")[0]
		return strings.TrimSpace(folderId)
	}

	// Extract folder ID from path
	path := parsedURL.Path
	parts := strings.Split(path, "/folder/")
	if len(parts) < 2 {
		return ""
	}

	// Get the folder ID (everything after /folder/ up to the next / or end)
	folderId := parts[1]
	// Remove any trailing path segments
	if idx := strings.Index(folderId, "/"); idx != -1 {
		folderId = folderId[:idx]
	}

	return strings.TrimSpace(folderId)
}

func main() {
	e := echo.New()
	e.POST("/get-links", getLinks)
	e.Logger.Fatal(e.Start(":8080"))
}
