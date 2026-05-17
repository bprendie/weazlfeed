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

func (c Client) Fetch(ctx context.Context, url string) (Feed, error) {
	if IsGopher(url) {
		return FetchGopher(ctx, url)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Feed{}, err
	}
	req.Header.Set("User-Agent", "weazlfeed/0.1")
	resp, err := c.http.Do(req)
	if err != nil {
		return Feed{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Feed{}, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return Feed{}, err
	}
	return Parse(body)
}

func IsGopher(raw string) bool {
	return strings.HasPrefix(strings.ToLower(raw), "gopher://")
}
