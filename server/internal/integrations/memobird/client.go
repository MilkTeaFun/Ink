package memobird

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	defaultBaseURL = "http://open.memobird.cn"
	defaultTimeout = 30 * time.Second
)

type Config struct {
	AccessKey  string
	DeviceID   string
	UserID     int
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	accessKey  string
	deviceID   string
	userID     atomic.Int64
}

type BaseResponse struct {
	ShowAPIResCode  int    `json:"showapi_res_code"`
	ShowAPIResError string `json:"showapi_res_error"`
}

type BindResponse struct {
	BaseResponse
	UserID int `json:"showapi_userid"`
}

type PrintResponse struct {
	BaseResponse
	Result         int    `json:"result"`
	SmartGuid      string `json:"smartGuid"`
	PrintContentID int    `json:"printcontentid"`
}

type PrintStatusResponse struct {
	BaseResponse
	PrintFlag      int `json:"printflag"`
	PrintContentID int `json:"printcontentid"`
}

type apiResponse interface {
	IsSuccess() bool
	Error() string
}

func NewClient(cfg Config) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		accessKey:  cfg.AccessKey,
		deviceID:   cfg.DeviceID,
	}
	client.SetUserID(cfg.UserID)
	return client
}

func (r *BaseResponse) IsSuccess() bool {
	return r.ShowAPIResCode == 1
}

func (r *BaseResponse) Error() string {
	return r.ShowAPIResError
}

func (p *PrintStatusResponse) IsPrinted() bool {
	return p.PrintFlag == 1
}

func (c *Client) BindAndRemember(ctx context.Context, userIdentifying string) (*BindResponse, error) {
	resp, err := c.BindUser(ctx, userIdentifying)
	if err != nil {
		return nil, err
	}
	c.SetUserID(resp.UserID)
	return resp, nil
}

func (c *Client) BindUser(ctx context.Context, userIdentifying string) (*BindResponse, error) {
	params := url.Values{}
	params.Set("memobirdID", c.deviceID)
	params.Set("useridentifying", userIdentifying)
	return postFormJSON[BindResponse](ctx, c, "/home/setuserbind", params, "bind")
}

func (c *Client) GetPrintStatus(ctx context.Context, printContentID int) (*PrintStatusResponse, error) {
	params := url.Values{}
	params.Set("printcontentid", fmt.Sprintf("%d", printContentID))
	return postFormJSON[PrintStatusResponse](ctx, c, "/home/getprintstatus", params, "get status")
}

func (c *Client) PrintHTML(ctx context.Context, html string) (*PrintResponse, error) {
	if c.GetUserID() == 0 {
		return nil, fmt.Errorf("user_id not configured")
	}

	encoded, err := encodeHTMLToGBKBase64(html)
	if err != nil {
		return nil, fmt.Errorf("failed to encode HTML: %w", err)
	}

	params := url.Values{}
	params.Set("memobirdID", c.deviceID)
	params.Set("userID", fmt.Sprintf("%d", c.GetUserID()))
	params.Set("printHtml", encoded)
	return postFormJSON[PrintResponse](ctx, c, "/home/printpaperFromHtml", params, "print from HTML")
}

func (c *Client) SetUserID(userID int) {
	c.userID.Store(int64(userID))
}

func (c *Client) GetUserID() int {
	return int(c.userID.Load())
}

func (c *Client) timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	params = cloneValues(params)
	params.Set("ak", c.accessKey)
	params.Set("timestamp", c.timestamp())

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", c.baseURL, endpoint),
		bytes.NewReader([]byte(params.Encode())),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return body, nil
}

func postFormJSON[T any](ctx context.Context, c *Client, endpoint string, params url.Values, action string) (*T, error) {
	body, err := c.doRequest(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var resp T
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response for %s: %w", action, err)
	}

	apiResp, ok := any(&resp).(apiResponse)
	if !ok {
		return nil, fmt.Errorf("response type does not implement apiResponse")
	}
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("%s failed: %s", action, apiResp.Error())
	}

	return &resp, nil
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, items := range values {
		cloned[key] = append([]string(nil), items...)
	}
	return cloned
}

func encodeHTMLToGBKBase64(html string) (string, error) {
	gbkBytes, _, err := transform.Bytes(simplifiedchinese.GBK.NewEncoder(), []byte(html))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gbkBytes), nil
}
