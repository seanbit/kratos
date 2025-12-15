package thirds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type FeiShuBot struct {
	WebhookURL string
}

// NewFeiShuBot creates a new FeiShuBot instance
func NewFeiShuBot(webhookURL string) *FeiShuBot {
	return &FeiShuBot{WebhookURL: webhookURL}
}

// SendTextMessage sends a text message to FeiShu
func (bot *FeiShuBot) SendTextMessage(text string) error {
	requestBody := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}

	return bot.sendPostRequest(requestBody)
}

// sendPostRequest sends a POST request to FeiShu
func (bot *FeiShuBot) sendPostRequest(requestBody interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("json Marshal failed: %v", err)
	}

	resp, err := http.Post(bot.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
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
