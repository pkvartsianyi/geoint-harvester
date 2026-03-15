package pinpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkvartsianyi/geoint-harvester/internal/domain"
)

type ExtractRequest struct {
	Text string `json:"text"`
}

type ExtractResponse struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type PinpointAdapter struct {
	httpClient *http.Client
	token      string
	url        string
}

func NewPinpointAdapter(token string) *PinpointAdapter {
	return &PinpointAdapter{
		httpClient: &http.Client{},
		token:      token,
		url:        "https://pinpoint-9clk.onrender.com/extract",
	}
}

func (a *PinpointAdapter) Extract(ctx context.Context, text string) (*domain.GeoPoint, error) {
	reqBody, err := json.Marshal(ExtractRequest{Text: text})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinpoint service returned status: %s", resp.Status)
	}

	var result ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &domain.GeoPoint{
		Type:        result.Type,
		Coordinates: result.Coordinates,
	}, nil
}
