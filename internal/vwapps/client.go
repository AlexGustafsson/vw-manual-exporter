package vwapps

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
)

type DigitalManualClient struct {
	Client   *http.Client
	Language string
}

func (c *DigitalManualClient) populateDefaultHeaders(req *http.Request) error {
	lang, market, ok := strings.Cut(c.Language, "_")
	if !ok {
		return fmt.Errorf("invalid language")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-fan-lang", lang)
	req.Header.Set("x-fan-market", market)

	return nil
}

type VehicleView string

const (
	VehicleViewSideLeft  VehicleView = "side-left"
	VehicleViewSideRight VehicleView = "side-right"
)

func (c *DigitalManualClient) GetVehiclePictureURL(ctx context.Context, vin string, view VehicleView) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://prod-a.dima.fan.vwapps.run/public/bff-digitalmanual/vehicle/picture/%s?vehicleView=%s", vin, view), nil)
	if err != nil {
		return "", err
	}

	if err := c.populateDefaultHeaders(req); err != nil {
		return "", err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response struct {
		PictureVIN  string `json:"pictureVin"` // Same as request
		PictureURL  string `json:"pictureUrl"`
		PictureType string `json:"pictureType"` // Typically "vilma"
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.PictureURL, nil
}

func (c *DigitalManualClient) GetTopicTree(ctx context.Context, vin string) (*TopicTree, error) {
	topicId, err := c.getDefaultEditionTopicID(ctx, vin)
	if err != nil {
		return nil, err
	}

	topicTree, err := c.getTopicTree(ctx, topicId, vin)
	if err != nil {
		return nil, err
	}

	return topicTree, nil
}

func (c *DigitalManualClient) GetTopicHTML(ctx context.Context, vin string, topicID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://userguide.volkswagen.de/default/api/consumer/V4/topic/%s", topicID), nil)
	if err != nil {
		return "", err
	}

	if err := c.populateDefaultHeaders(req); err != nil {
		return "", err
	}

	// Don't ask, I have no idea why they did this, but it's required by the API
	consumerToken := base64.RawURLEncoding.EncodeToString(fmt.Appendf([]byte{}, `{"vin":"%s","market":"RDW","tokenType":"vin","version":3}`, vin))
	req.Header.Set("X-Consumer-Token", consumerToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", ErrTopicNotFound
	} else if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response struct {
		TopicID  string `json:"topicId"`
		BodyHTML string `json:"bodyHtml"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.BodyHTML, nil
}

func (c *DigitalManualClient) getDefaultEditionTopicID(ctx context.Context, vin string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://prod-a.dima.fan.vwapps.run/public/bff-digitalmanual/digitalmanual/editions/default?vin=%s&language=%s&ownersManualTerms=Instruktionsbok", vin, c.Language), nil)
	if err != nil {
		return "", err
	}

	if err := c.populateDefaultHeaders(req); err != nil {
		return "", err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response struct {
		Results []struct {
			TopicID string `json:"topicId"`
			Title   string `json:"title"`
			// ... Category of unknown type
		} `json:"results"`
		ModelName      string `json:"modelName"`
		EditionVersion string `json:"editionVersion"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response.Results) == 0 {
		return "", fmt.Errorf("default edition has no results")
	}

	return response.Results[0].TopicID, nil
}

type TopicTree struct {
	ID       string        `json:"topicTreeId"`
	RootNode TopicTreeNode `json:"rootNode"`
}

type TopicTreeNode struct {
	NodeID        string          `json:"nodeId"`
	Title         string          `json:"title"`
	TargetTopicID string          `json:"targetTopicId,omitempty"`
	Children      []TopicTreeNode `json:"children"`
}

func (c *DigitalManualClient) getTopicTree(ctx context.Context, topicID string, vin string) (*TopicTree, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://prod-a.dima.fan.vwapps.run/public/bff-digitalmanual/digitalmanual/topictree/default/%s?vin=%s", topicID, vin), nil)
	if err != nil {
		return nil, err
	}

	if err := c.populateDefaultHeaders(req); err != nil {
		return nil, err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response TopicTree
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
