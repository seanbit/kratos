package thirds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultHTTPClient 默认 HTTP 客户端，带有合理的超时设置
var DefaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// PostJSON 发送 JSON POST 请求到指定 URL
// 返回响应体和错误
func PostJSON(url string, body interface{}) ([]byte, error) {
	return PostJSONWithClient(DefaultHTTPClient, url, body)
}

// PostJSONWithClient 使用指定的 HTTP 客户端发送 JSON POST 请求
func PostJSONWithClient(client *http.Client, url string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("http post failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	return respBody, nil
}
