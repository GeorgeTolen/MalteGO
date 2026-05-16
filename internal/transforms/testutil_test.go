package transforms

// Common test helpers shared across all transform test files in this package.

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/greynoise-maltego/maltego-go/internal/greynoise"
	"github.com/greynoise-maltego/maltego-go/internal/maltego"
)

// makeReq builds a minimal maltego.Request for a transform under test.
func makeReq(value string, settings ...map[string]string) *maltego.Request {
	s := map[string]string{"greynoise.api.key": "test-key"}
	if len(settings) > 0 {
		for k, v := range settings[0] {
			s[k] = v
		}
	}
	return &maltego.Request{
		Value:      value,
		EntityType: maltego.EntityIPv4Address,
		Properties: map[string]string{},
		Settings:   s,
		SoftLimit:  12,
		HardLimit:  12,
	}
}

// runTransform executes a transform with a mock client and returns parsed XML.
func runTransform(t *testing.T, tr Transform, client greynoise.Client, req *maltego.Request) parsedXML {
	t.Helper()
	resp, err := tr.Run(context.Background(), client, req)
	if err != nil {
		t.Fatalf("%s.Run returned unexpected error: %v", tr.Name(), err)
	}
	data, err := resp.ToXML()
	if err != nil {
		t.Fatalf("ToXML error: %v", err)
	}
	return mustParseXML(t, data)
}

// parsedXML is a lightweight view into a MaltegoTransformResponseMessage.
type parsedXML struct {
	Entities []parsedEntity
	Messages []parsedMsg
}

type parsedEntity struct {
	Type       string
	Value      string
	Properties map[string]string
}

type parsedMsg struct {
	Type  string
	Value string
}

func mustParseXML(t *testing.T, data []byte) parsedXML {
	t.Helper()

	type xmlField struct {
		Name  string `xml:"Name,attr"`
		Value string `xml:",chardata"`
	}
	type xmlEntity struct {
		Type             string `xml:"Type,attr"`
		Value            string `xml:"Value"`
		AdditionalFields *struct {
			Fields []xmlField `xml:"Field"`
		} `xml:"AdditionalFields"`
	}
	type xmlMsg struct {
		MessageType string `xml:"MessageType,attr"`
		Value       string `xml:",chardata"`
	}
	type xmlRoot struct {
		XMLName  xml.Name `xml:"MaltegoMessage"`
		Response struct {
			Entities struct {
				Entities []xmlEntity `xml:"Entity"`
			} `xml:"Entities"`
			UIMessages struct {
				Messages []xmlMsg `xml:"UIMessage"`
			} `xml:"UIMessages"`
		} `xml:"MaltegoTransformResponseMessage"`
	}

	raw := strings.TrimPrefix(string(data), xml.Header)
	var root xmlRoot
	if err := xml.Unmarshal([]byte(raw), &root); err != nil {
		t.Fatalf("parse response XML: %v\nXML:\n%s", err, string(data))
	}

	out := parsedXML{}
	for _, xe := range root.Response.Entities.Entities {
		pe := parsedEntity{
			Type:       xe.Type,
			Value:      xe.Value,
			Properties: map[string]string{},
		}
		if xe.AdditionalFields != nil {
			for _, f := range xe.AdditionalFields.Fields {
				pe.Properties[f.Name] = f.Value
			}
		}
		out.Entities = append(out.Entities, pe)
	}
	for _, m := range root.Response.UIMessages.Messages {
		out.Messages = append(out.Messages, parsedMsg{Type: m.MessageType, Value: m.Value})
	}
	return out
}

// assertFatalError verifies the response contains exactly one FatalError UIMessage.
func assertFatalError(t *testing.T, px parsedXML) {
	t.Helper()
	if len(px.Entities) != 0 {
		t.Errorf("expected no entities in error response, got %d", len(px.Entities))
	}
	if len(px.Messages) == 0 {
		t.Fatal("expected at least one UIMessage")
	}
	if px.Messages[0].Type != "FatalError" {
		t.Errorf("UIMessage type = %q, want FatalError", px.Messages[0].Type)
	}
}

// assertInform verifies the response contains at least one Inform UIMessage
// and no entities.
func assertInformNoEntities(t *testing.T, px parsedXML) {
	t.Helper()
	if len(px.Entities) != 0 {
		t.Errorf("expected no entities, got %d", len(px.Entities))
	}
	for _, m := range px.Messages {
		if m.Type == "Inform" {
			return
		}
	}
	t.Error("expected at least one Inform UIMessage")
}

// entityValues extracts all entity values from the parsed response.
func entityValues(px parsedXML) []string {
	out := make([]string, len(px.Entities))
	for i, e := range px.Entities {
		out[i] = e.Value
	}
	return out
}
