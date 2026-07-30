package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/terrasquid/terraform-provider-terrasquid/internal/model"
)

func (c *APIClient) ListDestinationGroups(ctx context.Context, name string) ([]model.DestinationGroup, error) {
	endpoint := "/api/v1/destination-groups/"
	if name != "" {
		endpoint += "?" + url.Values{"name": []string{name}}.Encode()
	}
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	var result []model.DestinationGroup
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) CreateDestinationGroup(ctx context.Context, input model.DestinationGroupInput) (*model.DestinationGroup, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/destination-groups/", input)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	var result model.DestinationGroup
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetDestinationGroup(ctx context.Context, id string) (*model.DestinationGroup, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/destination-groups/%s/", id), nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	var result model.DestinationGroup
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) UpdateDestinationGroup(ctx context.Context, id string, input model.DestinationGroupInput) (*model.DestinationGroup, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/destination-groups/%s/", id), input)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	var result model.DestinationGroup
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) DeleteDestinationGroup(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/destination-groups/%s/", id), nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return parseResponse(resp, nil)
}
