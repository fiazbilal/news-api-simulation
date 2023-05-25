package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	SimpleContentRequest = httptest.NewRequest("GET", "/?offset=0&count=5", nil)
	OffsetContentRequest = httptest.NewRequest("GET", "/?offset=5&count=5", nil)
)

func runRequest(t *testing.T, srv http.Handler, r *http.Request) (content []*ContentItem) {
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, r)

	if response.Code != 200 {
		t.Fatalf("Response code is %d, want 200", response.Code)
		return
	}

	decoder := json.NewDecoder(response.Body)
	err := decoder.Decode(&content)
	if err != nil {
		t.Fatalf("couldn't decode Response json: %v", err)
	}

	return content
}

func TestResponseCount(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

}

func TestResponseOrder(t *testing.T) {
	content := runRequest(t, app, SimpleContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for i, item := range content {
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

func TestOffsetResponseOrder(t *testing.T) {
	content := runRequest(t, app, OffsetContentRequest)

	if len(content) != 5 {
		t.Fatalf("Got %d items back, want 5", len(content))
	}

	for j, item := range content {
		i := j + 5
		if Provider(item.Source) != DefaultConfig[i%len(DefaultConfig)].Type {
			t.Errorf(
				"Position %d: Got Provider %v instead of Provider %v",
				i, item.Source, DefaultConfig[i].Type,
			)
		}
	}
}

// MockContentProvider is a mock provider that always fails to fetch content
type MockContentProvider struct {
	Source Provider
}

// GetContent always returns an error to simulate a failure
func (cp MockContentProvider) GetContent(userIP string, count int) ([]*ContentItem, error) {
	return nil, fmt.Errorf("MockContentProvider: Error fetching content")
}

func TestFallbackRespected(t *testing.T) {
	// Create a mock client that always fails to fetch content
	mockClient := &MockContentProvider{
		Source: Provider2,
	}

	// Replace the client for Provider2 with the mock client
	app.ContentClients[Provider2] = mockClient

	// Perform the request
	content := runRequest(t, app, SimpleContentRequest)

	// Verify the response
	if len(content) != 2 {
		t.Fatalf("Got %d items back, want 2", len(content))
	}

	// Verify that the response contains items only from Provider1
	for _, item := range content {
		if Provider(item.Source) != Provider1 {
			t.Errorf("Got Provider %v instead of Provider %v", item.Source, Provider1)
		}
	}
}
