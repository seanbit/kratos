package thirds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Alarm represents an alarm structure with necessary information.
type Alarm struct {
	ServiceName string `json:"service_name"`
	LarkURL     string `json:"lark_url"`
}

// NewAlarm creates a new Alarm instance.
func NewAlarm(serviceName string, larkURL string) *Alarm {
	return &Alarm{
		ServiceName: serviceName,
		LarkURL:     larkURL,
	}
}

type AlarmTextMessage struct {
	TraceId   string `json:"trace_id"`
	Operation string `json:"operation"`
	Title     string `json:"title"`
	Info      string `json:"info"`
}

func (a *Alarm) SendTextMessage(ctx context.Context, msg *AlarmTextMessage) error {

	lokiUrl := a.getLokiUrl(msg.TraceId)

	card := map[string]interface{}{
		"header": map[string]interface{}{
			"title": map[string]string{
				"content": msg.Title,
				"tag":     "plain_text",
			},
			"template": "blue",
		},
		"elements": []map[string]interface{}{
			{
				"tag": "div",
				"text": map[string]string{
					"content": fmt.Sprintf("Service Name: %s", a.ServiceName),
					"tag":     "plain_text",
				},
			},
			{
				"tag": "div",
				"text": map[string]string{
					"content": fmt.Sprintf("Info: %s", msg.Info),
					"tag":     "plain_text",
				},
			},
			{
				"tag": "div",
				"text": map[string]string{
					"content": fmt.Sprintf("Operation: %s", msg.Operation),
					"tag":     "plain_text",
				},
			},
			{
				"tag": "div",
				"text": map[string]string{
					"content": fmt.Sprintf("Request ID: %s", msg.TraceId),
					"tag":     "plain_text",
				},
			},
			{
				"tag": "action",
				"actions": []map[string]interface{}{
					{
						"tag": "button",
						"text": map[string]string{
							"tag":     "plain_text",
							"content": "Loki",
						},
						"url":  lokiUrl,
						"type": "primary",
					},
				},
			},
		},
	}

	requestBody := map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	}

	return a.pushToLark(requestBody)
}

func (a *Alarm) getLokiUrl(traceId string) string {
	lokiUrl := "https://grafana.carv.io/explore?schemaVersion=1&panes=%7B%22r8g%22:%7B%22datasource%22:%22loki%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22expr%22:%22%7Bnamespace%3D%5C%22carv-api%5C%22,%20app%3D%5C%22"
	lokiUrl += a.ServiceName
	lokiUrl += "%5C%22%7D%20%7C%3D%20%60"
	lokiUrl += traceId
	lokiUrl += "%60%22,%22queryType%22:%22range%22,%22datasource%22:%7B%22type%22:%22loki%22,%22uid%22:%22loki%22%7D,%22editorMode%22:%22builder%22,%22direction%22:%22backward%22%7D%5D,%22range%22:%7B%22from%22:%22now-30m%22,%22to%22:%22now%22%7D,%22panelsState%22:%7B%22logs%22:%7B%22visualisationType%22:%22logs%22%7D%7D%7D%7D&orgId=1"
	return lokiUrl
}

// PushToLark sends the alarm information to Lark.
func (a *Alarm) pushToLark(requestBody interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("json Marshal failed: %v", err)
	}

	resp, err := http.Post(a.LarkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("http Post failed: %v", resp.Status)
	}
}
