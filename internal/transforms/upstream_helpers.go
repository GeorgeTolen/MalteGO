package transforms

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

func addInputEntity(resp *maltego.Response, req *maltego.Request, defaultType string) *maltego.EntityBuilder {
	entityType := req.EntityType
	if entityType == "" {
		entityType = defaultType
	}
	e := resp.AddEntity(entityType, req.Value)
	for k, v := range req.Properties {
		e.AddProperty(k, "", maltego.MatchingRuleLoose, v)
	}
	return e
}

func addCommunityDisplayInfo(ipEnt *maltego.EntityBuilder, classification, lastSeen, link, name string) {
	linkText := ""
	if link != "" {
		linkText = fmt.Sprintf(`<h3><a href="%s">Open in GreyNoise</a></h3> <br/>`, link)
	}
	classificationText := ""
	if classification != "" {
		classificationText = "GreyNoise classification for IP: " + classification + "<br/>"
	}
	nameText := ""
	if name != "" && name != "unknown" {
		nameText = "GreyNoise attribution: " + name + "<br/>"
	}
	lastSeenText := ""
	if lastSeen != "" {
		lastSeenText = "Last seen by GreyNoise: " + lastSeen
	}
	ipEnt.AddDisplayInfo("GreyNoise Community", linkText+classificationText+nameText+lastSeenText)
	addClassificationOverlay(ipEnt, classification)
}

func addNoiseDisplayInfo(ipEnt *maltego.EntityBuilder, classification, lastSeen, link, name string, tags []string) {
	linkText := ""
	if link != "" {
		linkText = fmt.Sprintf(`<h3><a href="%s">Open in GreyNoise</a></h3> <br/>`, link)
	}
	classificationText := ""
	if classification != "" {
		classificationText = "GreyNoise classification for IP: " + classification + "<br/>"
	}
	nameText := ""
	if name != "" && name != "unknown" {
		nameText = "GreyNoise attribution: " + name + "<br/>"
	}
	lastSeenText := ""
	if lastSeen != "" {
		lastSeenText = "Last seen by GreyNoise: " + lastSeen + "<br/>"
	}
	tagText := ""
	if len(tags) > 0 {
		tagText = "GreyNoise Tags: " + strings.Join(tags, ", ")
	}
	ipEnt.AddDisplayInfo("GreyNoise", linkText+classificationText+nameText+lastSeenText+tagText)
	addClassificationOverlay(ipEnt, classification)
}

func addRIOTDisplayInfo(ipEnt *maltego.EntityBuilder, classification, lastUpdated, link, name string) {
	linkText := ""
	if link != "" {
		linkText = fmt.Sprintf(`<h3><a href="%s">Open in GreyNoise</a></h3> <br/>`, link)
	}
	classificationText := ""
	if classification != "" {
		classificationText = "GreyNoise classification for IP: " + classification + "<br/>"
	}
	nameText := ""
	if name != "" && name != "unknown" {
		nameText = "GreyNoise attribution: " + name + "<br/>"
	}
	lastUpdatedText := ""
	if lastUpdated != "" {
		lastUpdatedText = "Last updated by GreyNoise: " + lastUpdated
	}
	ipEnt.AddDisplayInfo("GreyNoise", linkText+classificationText+nameText+lastUpdatedText)
	addClassificationOverlay(ipEnt, classification)
}

func addClassificationOverlay(ipEnt *maltego.EntityBuilder, classification string) {
	colour := ""
	switch classification {
	case "benign":
		colour = "#45e06f"
	case "malicious":
		colour = "#eb4d4b"
	}
	if colour == "" {
		return
	}
	ipEnt.AddProperty("gn_color", "GreyNoise color", maltego.MatchingRuleLoose, colour).
		AddOverlay("gn_color", "NW", "colour")
}

func contextLink(r *greynoise.ContextResponse) string {
	if r.IP == "" {
		return ""
	}
	return "https://www.greynoise.io/viz/ip/" + r.IP
}

func intSetting(req *maltego.Request, name string, fallback int) int {
	if req.Settings == nil || req.Settings[name] == "" {
		return fallback
	}
	n, err := strconv.Atoi(req.Settings[name])
	if err != nil {
		return fallback
	}
	return n
}

func queryDateRange(req *maltego.Request) (string, string) {
	queryRange := ""
	if req.Settings != nil {
		queryRange = req.Settings["queryTimeRange"]
	}
	parts := strings.Split(queryRange, "-")
	if len(parts) == 2 {
		fromPart := strings.Split(parts[0], ".")[0]
		toPart := strings.Split(parts[1], ".")[0]
		fromTime, fromErr := strconv.ParseInt(fromPart, 10, 64)
		toTime, toErr := strconv.ParseInt(toPart, 10, 64)
		if fromErr == nil && toErr == nil {
			return time.Unix(fromTime, 0).Format("2006-01-02"), time.Unix(toTime, 0).Format("2006-01-02")
		}
	}
	now := time.Now().Format("2006-01-02")
	return now, now
}
