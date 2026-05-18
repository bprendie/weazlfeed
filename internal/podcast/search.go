package podcast

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const searchURL = "https://itunes.apple.com/search"

type Result struct {
	Title        string
	Author       string
	FeedURL      string
	Collection   string
	EpisodeCount int
}

type Client struct {
	http *http.Client
}

func NewClient() Client {
	return Client{http: &http.Client{Timeout: 20 * time.Second}}
}

func (c Client) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	params := url.Values{}
	params.Set("term", query)
	params.Set("media", "podcast")
	params.Set("entity", "podcast")
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("country", "US")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "weazlfeed/0.1")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("podcast search: %s", resp.Status)
	}
	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(payload.Results))
	for _, src := range payload.Results {
		if strings.TrimSpace(src.FeedURL) == "" {
			continue
		}
		results = append(results, Result{
			Title:        firstText(src.CollectionName, src.TrackName),
			Author:       src.ArtistName,
			FeedURL:      strings.TrimSpace(src.FeedURL),
			Collection:   src.CollectionViewURL,
			EpisodeCount: src.TrackCount,
		})
	}
	return results, nil
}

type searchResponse struct {
	Results []searchItem `json:"results"`
}

type searchItem struct {
	CollectionName    string `json:"collectionName"`
	TrackName         string `json:"trackName"`
	ArtistName        string `json:"artistName"`
	FeedURL           string `json:"feedUrl"`
	CollectionViewURL string `json:"collectionViewUrl"`
	TrackCount        int    `json:"trackCount"`
}

func firstText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
