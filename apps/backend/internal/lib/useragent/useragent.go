package useragent

import (
	"regexp"

	ua "github.com/mssola/useragent"
)

var botPattern = regexp.MustCompile(`(?i)Googlebot|Bingbot|Slackbot|Discordbot|Twitterbot|facebookexternalhit|WhatsApp`)

type Details struct {
	Browser        string
	BrowserVersion string
	OS             string
	Device         string
	IsBot          bool
}

func Parse(raw string) Details {
	p := ua.New(raw)

	browser, version := p.Browser()

	device := "desktop"

	if p.Mobile() {
		device = "mobile"
	}

	if p.Bot() {
		device = "bot"
	}

	return Details{
		Browser:        browser,
		BrowserVersion: version,
		OS:             p.OS(),
		Device:         device,
		IsBot:          p.Bot() || botPattern.MatchString(raw),
	}
}
