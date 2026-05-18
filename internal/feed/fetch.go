package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http *http.Client
}

func NewClient() Client {
	return Client{http: &http.Client{Timeout: 25 * time.Second}}
}

func (c Client) Fetch(ctx context.Context, url, etag, modified string) (Feed, error) {
	if IsGopher(url) {
		return FetchGopher(ctx, url)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Feed{}, err
	}
	req.Header.Set("User-Agent", "weazlfeed/0.1")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if modified != "" {
		req.Header.Set("If-Modified-Since", modified)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Feed{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return Feed{Status: resp.StatusCode, ETag: resp.Header.Get("ETag"), LastModified: resp.Header.Get("Last-Modified"), NotModified: true}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Feed{}, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return Feed{}, err
	}
	parsed, err := Parse(body)
	parsed.Status = resp.StatusCode
	parsed.ETag = resp.Header.Get("ETag")
	parsed.LastModified = resp.Header.Get("Last-Modified")
	return parsed, err
}

func IsGopher(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), "gopher://")
}
