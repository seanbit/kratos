package thirds

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

	_, err := PostJSON(bot.WebhookURL, requestBody)
	return err
}
