package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moov-io/ftdc-from-tap-to-auth/cardpersonalizer/models"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(baseURL string) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 5 * time.Second,
		},
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (i *Client) PersonalizeCard(req models.CardRequest) (models.CardResponse, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return models.CardResponse{}, err
	}

	res, err := i.httpClient.Post(i.baseURL+"/cards", "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return models.CardResponse{}, err
	}

	if res.StatusCode != http.StatusCreated {
		return models.CardResponse{}, fmt.Errorf("unexpected status code: %d; expected: %d", res.StatusCode, http.StatusCreated)
	}

	var card models.CardResponse
	err = json.NewDecoder(res.Body).Decode(&card)
	if err != nil {
		return models.CardResponse{}, err
	}

	return card, nil
}
