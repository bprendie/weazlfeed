package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bprendie/weazlfeed/internal/config"
)

type Client struct {
	provider config.Provider
	http     *http.Client
}

func New(provider config.Provider) Client {
	return Client{provider: provider, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c Client) Available(ctx context.Context) bool {
	if c.provider.ServerURL == "" || c.provider.Model == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.provider.ServerURL, "/"), nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}

func (c Client) Triage(ctx context.Context, markdown string) (string, error) {
	prompt := "Extract the three most critical technical points. No filler.\n\n" + markdown
	return c.Complete(ctx, prompt)
}

func (c Client) Ask(ctx context.Context, markdown, question string) (string, error) {
	prompt := "Answer using only this feed item. Be concise.\n\nARTICLE:\n" + markdown + "\n\nQUESTION:\n" + question
	return c.Complete(ctx, prompt)
}

func (c Client) FlagSludge(ctx context.Context, markdown string, rules []string) (bool, error) {
	if len(rules) == 0 {
		return false, nil
	}
	prompt := "Return only YES or NO. Should this item be flagged as SEO sludge?\nRules:\n- "
	prompt += strings.Join(rules, "\n- ") + "\n\nItem:\n" + markdown
	out, err := c.Complete(ctx, prompt)
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(out)), "YES"), err
}

func (c Client) Complete(ctx context.Context, prompt string) (string, error) {
	switch c.provider.Type {
	case "ollama":
		return c.ollama(ctx, prompt)
	case "vllm", "openai":
		return c.openAI(ctx, prompt)
	default:
		return "", errors.New("unknown provider")
	}
}

func (c Client) ollama(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{"model": c.provider.Model, "prompt": prompt, "stream": false}
	var out struct {
		Response string `json:"response"`
	}
	if err := c.post(ctx, "/api/generate", body, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Response), nil
}

func (c Client) openAI(ctx context.Context, prompt string) (string, error) {
	body := map[string]any{
		"model": c.provider.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := c.post(ctx, "/v1/chat/completions", body, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("provider returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func (c Client) post(ctx context.Context, path string, body any, out any) error {
	payload, _ := json.Marshal(body)
	url := strings.TrimRight(c.provider.ServerURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.provider.APIKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("provider %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
