package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const apifyRunSyncURL = "https://api.apify.com/v2/actor-tasks/ahmed_hrid~nike-product-scraper-task/run-sync-get-dataset-items"

// apifyDatasetItem maps one item from the nike-product-scraper-task output.
type apifyDatasetItem struct {
	ID         string `json:"id"`
	StyleColor string `json:"styleColor"`
	Title      string `json:"title"`
	Price      struct {
		Currency string  `json:"currency"`
		Current  float64 `json:"current"`
	} `json:"price"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// scrapeNikeProduct calls the Apify run-sync endpoint and returns the first
// scraped item. It blocks until the actor finishes (or ctx is cancelled).
func (app *application) scrapeNikeProduct(ctx context.Context, productURL string) (apifyDatasetItem, error) {
	endpoint := fmt.Sprintf("%s?token=%s", apifyRunSyncURL, url.QueryEscape(app.config.apify.token))

	body, err := json.Marshal(map[string]any{
		"startUrls": []map[string]string{{"url": productURL}},
	})
	if err != nil {
		return apifyDatasetItem{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return apifyDatasetItem{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return apifyDatasetItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return apifyDatasetItem{}, fmt.Errorf("apify scrape failed (%d): %s", resp.StatusCode, string(b))
	}

	var items []apifyDatasetItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return apifyDatasetItem{}, err
	}
	if len(items) == 0 {
		return apifyDatasetItem{}, fmt.Errorf("scraper returned no items for that URL")
	}

	return items[0], nil
}

// slugFromNikeURL extracts the slug from a Nike product URL.
// e.g. https://www.nike.com/t/air-max-dn8-leather-mens-shoes-GbnAW5Hb/IB6381-200
//
//	→ "air-max-dn8-leather-mens-shoes-GbnAW5Hb"
func slugFromNikeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	// path: t / {slug} / {styleColor}
	if len(parts) >= 3 && parts[0] == "t" {
		return parts[1]
	}
	return ""
}
