package helper

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}

func normalizeDomainValue(value string) string {
	newValue := strings.TrimSpace(value)
	newValue = strings.Replace(newValue, "http://", "", 1)
	newValue = strings.Replace(newValue, "https://", "", 1)
	newValue = strings.Replace(newValue, "www.", "", 1)
	newValue = strings.Split(newValue, "/")[0]
	return strings.TrimSuffix(newValue, "/")
}

func RemoveDomainError(url string) bool {
	serviceDomain := os.Getenv("DOMAIN")
	if serviceDomain == "" {
		serviceDomain = os.Getenv("RENDER_EXTERNAL_URL")
	}
	if serviceDomain == "" {
		return true
	}

	if normalizeDomainValue(url) == normalizeDomainValue(serviceDomain) {
		return false
	}

	return true
}

func ServiceBaseURL(fallback string) string {
	serviceDomain := os.Getenv("DOMAIN")
	if serviceDomain == "" {
		serviceDomain = os.Getenv("RENDER_EXTERNAL_URL")
	}
	if serviceDomain == "" {
		serviceDomain = fallback
	}
	return strings.TrimSuffix(serviceDomain, "/")
}
