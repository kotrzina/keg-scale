package ai

import (
	"fmt"
	"io"
	"net/http"
)

func ProviderLesempolemRegistered() (string, error) {
	const url = "https://lesempolem-backend-460110181987.europe-west1.run.app/lesempolem2025"

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get response from Lesempolem: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read body from Lesempolem: %w", err)
	}

	return string(body), nil
}
