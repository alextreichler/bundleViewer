package server

import (
	"encoding/json"
	"fmt"
	"html/template" // Added for logging Glob errors
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/alextreichler/bundleViewer/internal/cache" // Re-add cache import
	"github.com/alextreichler/bundleViewer/internal/models"
	"github.com/alextreichler/bundleViewer/internal/parser"
)

var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func getNodeHostname(bundlePath string, cachedData *cache.CachedData) string { // Changed to cachedData
	if bundlePath == "" || cachedData == nil {
		return "N/A"
	}
	unameInfo, err := parser.ParseUnameInfo(bundlePath)
	if err != nil {
		if cachedData != nil && len(cachedData.DataDiskFiles) > 0 {
			fileName := filepath.Base(cachedData.DataDiskFiles[0])
			parts := strings.Split(fileName, "_")
			if len(parts) >= 4 {
				return strings.TrimSuffix(parts[3], ".json")
			}
		}
		return "N/A"
	}
	return unameInfo.Hostname
}

func formatMilliseconds(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.2f ms", ms)
	}
	seconds := ms / 1000
	if seconds < 60 {
		return fmt.Sprintf("%.2f s", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		remainingSeconds := seconds - float64(int(minutes)*60)
		return fmt.Sprintf("%dm %.2fs", int(minutes), remainingSeconds)
	}
	hours := minutes / 60
	remainingMinutes := minutes - float64(int(hours)*60)
	return fmt.Sprintf("%dh %.2fm", int(hours), remainingMinutes)
}

func formatBytes(b float64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
		PB
		EB
	)

	switch {
	case b >= EB:
		return fmt.Sprintf("%.2f EB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2f PB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2f KB", b/KB)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// Helper function to traverse JSON path
func traverseJSONPath(data interface{}, jsonPath string) interface{} {
	if jsonPath == "" {
		return data
	}

	pathParts := strings.Split(jsonPath, ".")
	currentData := data

	for _, part := range pathParts {
		if part == "" {
			continue
		}

		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			idxStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil
			}
			if arr, ok := currentData.([]interface{}); ok && idx >= 0 && idx < len(arr) {
				currentData = arr[idx]
			} else {
				return nil
			}
		} else {
			if m, ok := currentData.(map[string]interface{}); ok {
				if val, exists := m[part]; exists {
					currentData = val
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	return currentData
}

// renderDataLazy renders data with lazy loading support for large structures
// startOffset is the index offset of the current data chunk within the original parent collection
func renderDataLazy(fileName string, jsonPath string, data interface{}, currentDepth int, startOffset int) template.HTML {
	var buf strings.Builder
	buf.Grow(1024)

	if currentDepth > maxInitialRenderDepth {
		switch data.(type) {
		case map[string]interface{}:
			return template.HTML("{...}")
		case []interface{}:
			return template.HTML("[...]")
		default:
			// Sanitize
			s := fmt.Sprintf("%v", data)
			return template.HTML(template.HTMLEscapeString(s))
		}
	}

	switch v := data.(type) {
	case []models.ClusterConfigEntry:
		buf.WriteString("<table><thead><tr><th>Config (key)</th><th>Value</th><th>Doc Link</th></tr></thead><tbody>")
		for _, entry := range v {
			buf.WriteString("<tr>")
			buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(entry.Key)))
			value := entry.Value
			if f, ok := value.(float64); ok {
				if strings.Contains(entry.Key, "bytes") || strings.Contains(entry.Key, "size") {
					buf.WriteString(fmt.Sprintf("<td>%s (%.0f B)</td>", formatBytes(f), f))
				} else if strings.Contains(entry.Key, "ms") {
					buf.WriteString(fmt.Sprintf("<td>%s (%.0f ms)</td>", formatMilliseconds(f), f))
				} else {
					s := strconv.FormatFloat(f, 'f', -1, 64)
					buf.WriteString(fmt.Sprintf("<td>%s</td>", s))
				}
			} else {
				buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(fmt.Sprintf("%v", value))))
			}
			buf.WriteString(fmt.Sprintf("<td><a href=\"%s\" target=\"_blank\">doc</a></td>", template.HTMLEscapeString(entry.DocLink)))
			buf.WriteString("</tr>")
		}
		buf.WriteString("</tbody></table>")

	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buf.WriteString("<table class=\"sortable-table\"><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody>")
		count := 0
		for _, k := range keys {
			if count >= maxInitialRenderItems {
				// HTMX Lazy Load Button
				// This replaces the current row with new rows (outerHTML)
				// We target 'this' (the button), but actually we want to append rows.
				// Table structure makes this tricky. We put the button in a TR.
				// If we swap outerHTML of the TR, we replace the button-row with the new data-rows.
				// AND the new data-rows might include a NEW button-row at the end.
				// URL params: offset = startOffset + count.
				nextOffset := startOffset + count
				url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
					template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

				buf.WriteString(fmt.Sprintf(`<tr><td colspan="2"><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest tr" hx-swap="outerHTML">Load %d more items...</button></td></tr>`,
					url, len(v)-count))
				break
			}
			buf.WriteString("<tr>")
			buf.WriteString(fmt.Sprintf("<td>%s</td>", template.HTMLEscapeString(k)))
			buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, fmt.Sprintf("%s.%s", jsonPath, k), v[k], currentDepth+1, 0)))
			buf.WriteString("</tr>")
			count++
		}
		buf.WriteString("</tbody></table>")

	case []interface{}:
		if len(v) == 0 {
			return template.HTML("[]")
		}

		// Try to render as table if elements are objects
		var processedSlice []map[string]interface{}
		for _, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				processedSlice = append(processedSlice, obj)
			} else {
				jsonBytes, err := json.Marshal(item)
				if err == nil {
					var objMap map[string]interface{}
					if json.Unmarshal(jsonBytes, &objMap) == nil {
						processedSlice = append(processedSlice, objMap)
					} else {
						processedSlice = nil
						break
					}
				} else {
					processedSlice = nil
					break
				}
			}
		}

		if len(processedSlice) > 0 {
			headersMap := make(map[string]struct{})
			for _, item := range processedSlice {
				for k := range item {
					headersMap[k] = struct{}{}
				}
			}

			var headers []string
			var otherHeaders []string

			if _, ok := headersMap["Key"]; ok {
				headers = append(headers, "Key")
			}
			if _, ok := headersMap["Value"]; ok {
				headers = append(headers, "Value")
			}
			if _, ok := headersMap["DocLink"]; ok {
				headers = append(headers, "DocLink")
			}

			for h := range headersMap {
				isPredefined := false
				switch h {
				case "Key", "Value", "DocLink":
					isPredefined = true
				}
				if !isPredefined {
					otherHeaders = append(otherHeaders, h)
				}
			}
			sort.Strings(otherHeaders)
			headers = append(headers, otherHeaders...)

			buf.WriteString("<table class=\"sortable-table\"><thead><tr>")
			for _, h := range headers {
				displayHeader := template.HTMLEscapeString(h)
				switch h {
				case "Key":
					displayHeader = "Config (key)"
				case "DocLink":
					displayHeader = "Doc Link"
				}
				buf.WriteString(fmt.Sprintf("<th>%s</th>", displayHeader))
			}
			buf.WriteString("</tr></thead><tbody>")

			for i, item := range processedSlice {
				if i >= maxInitialRenderItems {
					nextOffset := startOffset + i
					url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
						template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

					buf.WriteString(fmt.Sprintf(`<tr><td colspan="%d"><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest tr" hx-swap="outerHTML">Load %d more items...</button></td></tr>`,
						len(headers), url, len(processedSlice)-i))
					break
				}

				buf.WriteString("<tr>")
				for _, h := range headers {
					val := item[h]
					// Correctly calculate the absolute index for the child path
					absoluteIndex := startOffset + i
					childPath := fmt.Sprintf("%s[%d].%s", jsonPath, absoluteIndex, h)

					if h == "DocLink" {
						if s, ok := val.(string); ok && strings.HasPrefix(s, "http") {
							buf.WriteString(fmt.Sprintf("<td><a href=\"%s\" target=\"_blank\">doc</a></td>", template.HTMLEscapeString(s)))
						} else {
							buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, childPath, val, currentDepth+1, 0)))
						}
					} else {
						buf.WriteString(fmt.Sprintf("<td>%s</td>", renderDataLazy(fileName, childPath, val, currentDepth+1, 0)))
					}
				}
				buf.WriteString("</tr>")
			}
			buf.WriteString("</tbody></table>")
		} else {
			// Simple array
			buf.WriteString("<ul>")
			count := 0
			for i, item := range v {
				if count >= maxInitialRenderItems {
					nextOffset := startOffset + count
					url := fmt.Sprintf("/api/lazy-load?fileName=%s&jsonPath=%s&offset=%d&limit=%d",
						template.HTMLEscapeString(fileName), template.HTMLEscapeString(jsonPath), nextOffset, maxChunkSize)

					buf.WriteString(fmt.Sprintf(`<li><button class="lazy-load-btn" hx-get="%s" hx-trigger="click" hx-target="closest li" hx-swap="outerHTML">Load %d more items...</button></li>`,
						url, len(v)-count))
					break
				}
				// Correctly calculate the absolute index for the child path
				absoluteIndex := startOffset + i
				childPath := fmt.Sprintf("%s[%d]", jsonPath, absoluteIndex)

				buf.WriteString(fmt.Sprintf("<li>%s</li>", renderDataLazy(fileName, childPath, item, currentDepth+1, 0)))
				count++
			}
			buf.WriteString("</ul>")
		}

	default:
		if s, ok := data.(string); ok {
			buf.WriteString("<pre><code>")
			// Truncate very long strings
			if len(s) > 10000 {
				buf.WriteString(template.HTMLEscapeString(s[:10000]))
				buf.WriteString("\n... (truncated)")
			} else {
				buf.WriteString(template.HTMLEscapeString(s))
			}
			buf.WriteString("</code></pre>")
			return template.HTML(buf.String())
		}

		if f, ok := data.(float64); ok {
			if fileName != "" && (strings.HasSuffix(fileName, "_bytes") || strings.HasSuffix(fileName, "_size")) {
				return template.HTML(fmt.Sprintf("%s (%.0f B)", formatBytes(f), f))
			}
			s := strconv.FormatFloat(f, 'f', -1, 64)
			return template.HTML(s)
		}
		s := fmt.Sprintf("%v", data)
		return template.HTML(template.HTMLEscapeString(s))
	}

	return template.HTML(buf.String())
}
