package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultProviderType = "openai-compatible"
	DefaultProviderName = "OpenAI Compatible"
	DefaultModel        = "gpt-4.1-mini"
)

type RuntimeConfig struct {
	ProviderName string
	ProviderType string
	BaseURL      string
	Model        string
	APIKey       string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ReplyResult struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	ProviderName string `json:"providerName"`
}

type CompletionClient interface {
	CreateReply(ctx context.Context, cfg RuntimeConfig, messages []ChatMessage) (ReplyResult, error)
}

type OpenAIClient struct {
	httpClient              *http.Client
	allowInsecurePrivateURL bool
}

func NewOpenAIClient(timeout time.Duration, allowInsecurePrivateURL bool) *OpenAIClient {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}

	return &OpenAIClient{
		httpClient:              &http.Client{Timeout: timeout},
		allowInsecurePrivateURL: allowInsecurePrivateURL,
	}
}

func (c *OpenAIClient) CreateReply(ctx context.Context, cfg RuntimeConfig, messages []ChatMessage) (ReplyResult, error) {
	if err := ValidateBaseURL(cfg.BaseURL, c.allowInsecurePrivateURL); err != nil {
		return ReplyResult{}, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return ReplyResult{}, ErrNotConfigured
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return ReplyResult{}, ErrInvalidConfig
	}
	if len(messages) == 0 {
		return ReplyResult{}, ErrInvalidInput
	}

	payload := map[string]any{
		"model":       cfg.Model,
		"messages":    messages,
		"temperature": 0.7,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ReplyResult{}, err
	}

	endpoint := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ReplyResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ReplyResult{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return ReplyResult{}, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return ReplyResult{}, fmt.Errorf("%w: upstream returned %d", ErrProviderUnavailable, resp.StatusCode)
	}

	var decoded struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ReplyResult{}, fmt.Errorf("%w: invalid provider response", ErrProviderUnavailable)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return ReplyResult{}, fmt.Errorf("%w: provider returned no reply", ErrProviderUnavailable)
	}

	model := strings.TrimSpace(decoded.Model)
	if model == "" {
		model = cfg.Model
	}

	return ReplyResult{
		Content:      strings.TrimSpace(decoded.Choices[0].Message.Content),
		Model:        model,
		ProviderName: chooseString(strings.TrimSpace(cfg.ProviderName), DefaultProviderName),
	}, nil
}

func ValidateBaseURL(raw string, allowInsecurePrivateURL bool) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ErrInvalidConfig
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidConfig
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return ErrInvalidConfig
	}
	if parsed.Scheme != "https" && (!allowInsecurePrivateURL || parsed.Scheme != "http") {
		return ErrInvalidConfig
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return ErrInvalidConfig
	}
	if !allowInsecurePrivateURL && strings.EqualFold(host, "localhost") {
		return ErrInvalidConfig
	}

	if addr, err := netip.ParseAddr(host); err == nil {
		if !allowInsecurePrivateURL && (addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast()) {
			return ErrInvalidConfig
		}
		return nil
	}

	if !allowInsecurePrivateURL {
		ips, err := net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				addr, parseErr := netip.ParseAddr(ip.String())
				if parseErr != nil {
					continue
				}
				if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() {
					return ErrInvalidConfig
				}
			}
		}
	}

	return nil
}

func chooseString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
