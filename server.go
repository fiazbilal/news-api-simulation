package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// App represents the server's internal state.
// It holds configuration about providers and content.
type App struct {
	ContentClients map[Provider]Client
	Config         ContentMix
}

func (app App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())

	// Parse URL parameters.
	qStr := req.URL.Query()
	countStr := qStr.Get("count")
	offsetStr := qStr.Get("offset")

	count, err := strconv.Atoi(countStr)
	if err != nil {
		http.Error(w, "failed to parse count param", http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		http.Error(w, "failed to parse offset param", http.StatusBadRequest)
		return
	}

	// Fetch content items.
	contentItems := make([]ContentItem, 0, count)

	config := app.Config
	configLen := len(config)

	for i := 0; i < count; i++ {
		index := (i + offset) % configLen
		contentConfig := config[index]

		provider := contentConfig.Type
		client := app.ContentClients[provider]

		// Fetch 1 content item at a time.
		item, err := client.GetContent("", 1)
		if err != nil {
			log.Printf("failed fetching content from provider %v: %v", provider, err)
			break
		}

		if len(item) > 0 {
			contentItems = append(contentItems, *item[0])
		}
	}

	// Encode content items as JSON.
	jsonData, err := json.Marshal(contentItems)
	if err != nil {
		http.Error(w, "failed encoding response", http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
