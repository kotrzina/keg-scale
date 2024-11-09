package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kotrzina/keg-scale/pkg/utils"
)

type Discord struct {
	hookURLs struct {
		open string
		keg  string
	}
}

func New(openHook, kegHook string) *Discord {
	return &Discord{
		hookURLs: struct {
			open string
			keg  string
		}{
			open: openHook,
			keg:  kegHook,
		},
	}
}

func (d *Discord) SendOpen() error {
	now := time.Now()
	utils.FormatDate(now)
	message := "🍻	**Hospoda otevřena!**"
	return d.sendWebhook(d.hookURLs.open, message)
}

func (d *Discord) SendKeg(keg int) error {
	message := fmt.Sprintf("🛢	**Naražena nová bečka:** %d l", keg)
	return d.sendWebhook(d.hookURLs.keg, message)
}

func (d *Discord) sendWebhook(url, message string) error {
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
