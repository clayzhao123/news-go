package crawler

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"news-go/internal/news"
)

type RSSFetcher struct {
	client *http.Client
}

func NewRSSFetcher(timeout time.Duration) *RSSFetcher {
	return &RSSFetcher{client: &http.Client{Timeout: timeout}}
}

type rssDocument struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (f *RSSFetcher) Fetch(ctx context.Context, feedURL string) ([]news.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("rss status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var doc rssDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	out := make([]news.Article, 0, len(doc.Channel.Items))
	for _, it := range doc.Channel.Items {
		published := time.Now().UTC()
		if t, err := time.Parse(time.RFC1123Z, it.PubDate); err == nil {
			published = t.UTC()
		} else if t, err := time.Parse(time.RFC1123, it.PubDate); err == nil {
			published = t.UTC()
		}
		out = append(out, news.Article{
			Title:       it.Title,
			URL:         it.Link,
			Source:      "rss",
			Content:     it.Description,
			PublishedAt: published,
		})
	}
	return out, nil
}
