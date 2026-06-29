// Package evolution is a thin client for the Evolution WhatsApp API, used by the
// booking bot to send outbound text messages. Each call is authenticated with
// the per-instance API key (tenant_channels.external_key), so no global key is
// involved.
package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client sends WhatsApp messages through an Evolution instance.
type Client interface {
	// SendText posts a text message to `to` (a WhatsApp JID/number) via the named
	// instance, authenticated with that instance's API key.
	SendText(ctx context.Context, instance, apiKey, to, text string) error
}

type httpClient struct {
	baseURL string
	http    *http.Client
}

// NewClient builds an Evolution client targeting baseURL (EVO_API_URL).
func NewClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *httpClient) SendText(ctx context.Context, instance, apiKey, to, text string) error {
	body, err := json.Marshal(map[string]string{"number": to, "text": text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/message/sendText/"+instance, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("evolution sendText failed: status %d: %s", resp.StatusCode, string(data))
	}
	return nil
}
