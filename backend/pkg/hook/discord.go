package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Discord struct {
	hookURLs struct {
		open string
		keg  string
	}

	ctx    context.Context
	logger *logrus.Logger
}

func New(ctx context.Context, openHook, kegHook string, logger *logrus.Logger) *Discord {
	return &Discord{
		hookURLs: struct {
			open string
			keg  string
		}{
			open: openHook,
			keg:  kegHook,
		},
		ctx:    ctx,
		logger: logger,
	}
}

func (d *Discord) SendOpen() {
	go func() {
		message := "🍻	**Hospoda otevřena!**"
		err := d.sendWebhook(d.hookURLs.open, message)
		if err != nil {
			d.logger.Errorf("could not send Discord webhook: %v", err)
		}
	}()
}

func (d *Discord) SendKeg(keg int) {
	go func() {
		message := fmt.Sprintf("🛢	**Naražena nová bečka:** %d l", keg)
		err := d.sendWebhook(d.hookURLs.keg, message)
		if err != nil {
			d.logger.Errorf("could not send Discord webhook: %v", err)
		}
	}()
}

func (d *Discord) sendWebhook(url, message string) error {
	if url == "" {
		return fmt.Errorf("Discord webhook URL is not set")
	}

	body := struct {
		Content string `json:"content"`
	}{
		Content: message,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("could not marshal data for Discord webhook")
	}
	data := bytes.NewBuffer(jsonData)

	req, err := http.NewRequestWithContext(d.ctx, http.MethodPost, url, data)
	if err != nil {
		return fmt.Errorf("could not create request for Discord webhook: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send Discord webhook: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("invalid response code from Discord webhook: %d", resp.StatusCode)
	}

	return nil
}
