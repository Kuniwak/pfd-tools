package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type WebhookClient struct {
	webhookURL *url.URL
	h          *http.Client
}

func NewWebhookClient(webhookURL *url.URL) *WebhookClient {
	return &WebhookClient{webhookURL: webhookURL, h: http.DefaultClient}
}

func (c *WebhookClient) Do(req *http.Request) error {
	resp, err := c.h.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack.WebhookClient.Do: status code is not OK: %d", resp.StatusCode)
	}

	return nil
}

type Message struct {
	Blocks []Block `json:"blocks"`
}

type Block interface {
	GetType() string
}

type SectionBlock struct {
	Text      Block `json:"text"`
	Accessory Block `json:"accessory,omitempty"`
}

var _ Block = &SectionBlock{}

func (b *SectionBlock) GetType() string {
	return "section"
}

type MarkdownBlock struct {
	Text string `json:"text"`
}

var _ Block = &MarkdownBlock{}

func (b *MarkdownBlock) GetType() string {
	return "mrkdwn"
}

type ButtonBlock struct {
	Text     Block  `json:"text"`
	URL      string `json:"url"`
	Value    string `json:"value"`
	ActionID string `json:"action_id"`
}

var _ Block = &ButtonBlock{}

func (b *ButtonBlock) GetType() string {
	return "button"
}

type ContextBlock struct {
	Elements []Block `json:"elements"`
}

var _ Block = &ContextBlock{}

func (b *ContextBlock) GetType() string {
	return "context"
}

type DividerBlock struct {
}

var _ Block = &DividerBlock{}

func (b *DividerBlock) GetType() string {
	return "divider"
}

type PostIncomingWebhookFunc func(message *Message) error

func NewPostIncomingWebhookFunc(c *WebhookClient) PostIncomingWebhookFunc {
	return func(message *Message) error {
		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, c.webhookURL.String(), bytes.NewReader(body))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		if err := c.Do(req); err != nil {
			return err
		}

		return nil
	}
}
