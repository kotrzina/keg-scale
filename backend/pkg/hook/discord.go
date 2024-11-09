package hook

import (
	"bytes"
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

	logger *logrus.Logger
}

func New(openHook, kegHook string, logger *logrus.Logger) *Discord {
	return &Discord{
		hookURLs: struct {
			open string
			keg  string
		}{
			open: openHook,
			keg:  kegHook,
		},
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

	resp, err := http.Post(url, "application/json", data)
	if err != nil {
		return fmt.Errorf("could not send Discord webhook: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("invalid response code from Discord webhook: %d", resp.StatusCode)
	}

	return nil
}
