package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/alextreichler/bundleViewer/internal/models"
)

type SearchResult struct {
	Category string
	Title    string
	Link     string
	Preview  template.HTML // Allow HTML for highlighting
}

type GlobalSearchPageData struct {
	BasePageData
	Query        string
	Results      []SearchResult
	NodeHostname string
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" || s.store == nil {
		pageData := GlobalSearchPageData{
			BasePageData: s.newBasePageData("Search"),
			NodeHostname: s.nodeHostname,
		}

		buf := builderPool.Get().(*strings.Builder)
		buf.Reset()
		defer builderPool.Put(buf)

		if err := s.searchTemplate.Execute(buf, pageData); err != nil {
			s.logger.Error("Failed to execute search template", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := io.WriteString(w, buf.String()); err != nil {
			s.logger.Error("Failed to write search response", "error", err)
		}
		return
	}

	var results []SearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	lowercaseQuery := strings.ToLower(query)

	// 1. Search Logs (Limit to top 50 matches)
	wg.Add(1)
	go func() {
		defer wg.Done()
		filter := models.LogFilter{
			Search: query, // Store handles FTS
			Limit:  50,
		}
		logs, _, err := s.store.GetLogs(&filter)
		if err == nil && len(logs) > 0 {
			mu.Lock()
			for _, log := range logs {
				var preview template.HTML
				if log.Snippet != "" {
					// Use DB-provided snippet with highlighting
					preview = template.HTML(log.Snippet)
				} else {
					// Fallback to manual highlighting
					preview = highlightMatch(log.Message, query)
				}

				results = append(results, SearchResult{
					Category: "Logs",
					Title:    fmt.Sprintf("%s [%s]", log.Timestamp.Format("15:04:05"), log.Level),
					Link:     fmt.Sprintf("/logs?search=%s", query), // Link to logs page with filter
					Preview:  preview,
				})
			}
			mu.Unlock()
		}
	}()

	// 2. Search In-Memory Files (Configs, Proc, Utils)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for group, files := range s.cachedData.GroupedFiles {
			for _, file := range files {
				contentStr := ""
				// Handle string or []byte data
				if str, ok := file.Data.(string); ok {
					contentStr = str
				} else if bytes, ok := file.Data.([]byte); ok {
					contentStr = string(bytes)
				} else if lines, ok := file.Data.([]string); ok {
					contentStr = strings.Join(lines, "\n")
				} else if file.Data != nil {
					// Fallback to JSON representation for structured data
				b, _ := json.Marshal(file.Data)
					contentStr = string(b)
				}

				if contentStr != "" {
					// Check for match
					if idx := strings.Index(strings.ToLower(contentStr), lowercaseQuery); idx != -1 {
						// Extract matching context
						start := idx - 50
						if start < 0 {
							start = 0
						}
						end := idx + len(query) + 50
						if end > len(contentStr) {
							end = len(contentStr)
						}
						
						snippet := contentStr[start:end]
						preview := highlightMatch(snippet, query)

						mu.Lock()
						results = append(results, SearchResult{
							Category: fmt.Sprintf("File: %s", group),
							Title:    file.FileName,
							Link:     fmt.Sprintf("/?file=%s&q=%s", file.FileName, query),
							Preview:  preview,
						})
						mu.Unlock()
					}
				}
			}
		}
	}()

	// 3. Search K8s Resources
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Helper to search generic map/structs could be useful here, but let's do manual checks for now
		// Pods
		for _, pod := range s.cachedData.K8sStore.Pods {
			if strings.Contains(strings.ToLower(pod.Metadata.Name), lowercaseQuery) {
				mu.Lock()
				results = append(results, SearchResult{
					Category: "Kubernetes",
					Title:    fmt.Sprintf("Pod: %s", pod.Metadata.Name),
					Link:     "/k8s",
					Preview:  template.HTML(fmt.Sprintf("Found Pod named <b>%s</b>", pod.Metadata.Name)),
				})
				mu.Unlock()
			}
		}
		// Services
		for _, svc := range s.cachedData.K8sStore.Services {
			if strings.Contains(strings.ToLower(svc.Metadata.Name), lowercaseQuery) {
				mu.Lock()
				results = append(results, SearchResult{
					Category: "Kubernetes",
					Title:    fmt.Sprintf("Service: %s", svc.Metadata.Name),
					Link:     "/k8s",
					Preview:  template.HTML(fmt.Sprintf("Found Service named <b>%s</b>", svc.Metadata.Name)),
				})
				mu.Unlock()
			}
		}
	}()

	wg.Wait()

	pageData := GlobalSearchPageData{
		BasePageData: s.newBasePageData("Search"),
		Query:        query,
		Results:      results,
		NodeHostname: s.nodeHostname,
	}

	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)

	if err := s.searchTemplate.Execute(buf, pageData); err != nil {
		s.logger.Error("Failed to execute search template with results", "error", err)
		http.Error(w, "Failed to render search results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.WriteString(w, buf.String()); err != nil {
		s.logger.Error("Failed to write search response", "error", err)
	}
}

func highlightMatch(text, query string) template.HTML {
	// Simple replace ignoring case is hard without regex
	// For now, simple case-insensitive replace
	// Note: Strings.Replace is case-sensitive.
	
	// Quick hack for case-insensitive highlighting
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	
	var sb strings.Builder
	lastIdx := 0
	
	for {
		idx := strings.Index(lowerText[lastIdx:], lowerQuery)
		if idx == -1 {
			sb.WriteString(template.HTMLEscapeString(text[lastIdx:]))
			break
		}
		
		idx += lastIdx
		sb.WriteString(template.HTMLEscapeString(text[lastIdx:idx]))
		sb.WriteString("<mark>")
		// The query string itself must also be escaped, even though it matches our search
		// This prevents XSS if the user searches for "<script>" and we blindly insert it back
		sb.WriteString(template.HTMLEscapeString(text[idx : idx+len(query)]))
		sb.WriteString("</mark>")
		lastIdx = idx + len(query)
	}
	
	return template.HTML(sb.String())
}
