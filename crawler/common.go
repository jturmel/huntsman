package crawler

import "strings"

// DetermineKind infers the resource kind from the Content-Type header
func DetermineKind(contentType string) string {
	if strings.Contains(contentType, "text/html") {
		return "document"
	} else if strings.Contains(contentType, "text/css") {
		return "stylesheet"
	} else if strings.Contains(contentType, "javascript") {
		return "script"
	} else if strings.Contains(contentType, "font") {
		return "font"
	} else if strings.Contains(contentType, "image/png") {
		return "png"
	} else if strings.Contains(contentType, "image/gif") {
		return "gif"
	} else if strings.Contains(contentType, "image/jpeg") {
		return "jpeg"
	} else if strings.Contains(contentType, "image/svg+xml") {
		return "svg+xml"
	} else if strings.Contains(contentType, "x-icon") || strings.Contains(contentType, "vnd.microsoft.icon") {
		return "x-icon"
	} else if strings.Contains(contentType, "manifest+json") {
		return "manifest"
	}
	return "Other"
}
