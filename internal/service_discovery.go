package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	discoveryPath = "/.well-known/terraform.json"
)

// TFEServiceDiscovery wraps a map from service name to URL.
type TFEServiceDiscovery struct {
	serviceMap map[string]interface{}
}

// DiscoverTFEServices retrieves the service discovery document and returns a map from service name to the associated URL.
func DiscoverTFEServices(httpClient *http.Client, endpoint *url.URL) (*TFEServiceDiscovery, error) {

	// Use the regular HTTP client to retrieve the discovery document.
	discoveryURL := url.URL{
		Scheme: endpoint.Scheme,
		Host:   endpoint.Host,
		Path:   discoveryPath,
	}
	resp, err := httpClient.Get(discoveryURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get TFE discovery document: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get TFE discovery document, status code: %d", resp.StatusCode)
	}
	discoveredBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TFE discovery document body: %w", err)
	}
	var discoveredServices map[string]interface{}
	err = json.Unmarshal(discoveredBytes, &discoveredServices)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TFE discovery document body: %w", err)
	}

	return &TFEServiceDiscovery{serviceMap: discoveredServices}, nil
}

// TODO: If a need is found for a discovered service other than a URL, a new method could go here.

// GetServiceURL returns the URL for the specified service.
func (t *TFEServiceDiscovery) GetServiceURL(key string) (*url.URL, error) {

	// Look up the key in the service map.
	val, ok := t.serviceMap[key]
	if !ok {
		return nil, fmt.Errorf("service not found for key %s", key)
	}

	// Verify/convert to a string.
	s, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("value for service key %s is not a string", key)
	}

	// Parse to a URL.
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service URL: %w", err)
	}

	return u, nil
}
