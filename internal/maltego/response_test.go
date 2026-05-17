package maltego

import (
	"encoding/xml"
	"strings"
	"testing"
)

// --- XML structure for asserting response content ---

type parsedResponse struct {
	XMLName  xml.Name `xml:"MaltegoMessage"`
	Response struct {
		Entities struct {
			Entities []struct {
				Type    string `xml:"Type,attr"`
				Value   string `xml:"Value"`
				Weight  int    `xml:"Weight"`
				IconURL string `xml:"IconURL"`
				DisplayInformation *struct {
					Labels []struct {
						Name    string `xml:"Name,attr"`
						Content string `xml:",innerxml"`
					} `xml:"Label"`
				} `xml:"DisplayInformation"`
				AdditionalFields *struct {
					Fields []struct {
						Name         string `xml:"Name,attr"`
						DisplayName  string `xml:"DisplayName,attr"`
						MatchingRule string `xml:"MatchingRule,attr"`
						Value        string `xml:",chardata"`
					} `xml:"Field"`
				} `xml:"AdditionalFields"`
			} `xml:"Entity"`
		} `xml:"Entities"`
		UIMessages struct {
			Messages []struct {
				MessageType string `xml:"MessageType,attr"`
				Value       string `xml:",chardata"`
			} `xml:"UIMessage"`
		} `xml:"UIMessages"`
	} `xml:"MaltegoTransformResponseMessage"`
}

func parseXMLResponse(t *testing.T, data []byte) parsedResponse {
	t.Helper()
	body := strings.TrimPrefix(string(data), xml.Header)
	var pr parsedResponse
	if err := xml.Unmarshal([]byte(body), &pr); err != nil {
		t.Fatalf("failed to parse response XML: %v\nXML:\n%s", err, string(data))
	}
	return pr
}

// --- Response builder tests ---

func TestNewResponse_ToXML_EmptyResponse(t *testing.T) {
	r := NewResponse()
	data, err := r.ToXML()
	if err != nil {
		t.Fatalf("ToXML error: %v", err)
	}

	pr := parseXMLResponse(t, data)

	if len(pr.Response.Entities.Entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(pr.Response.Entities.Entities))
	}
	if len(pr.Response.UIMessages.Messages) != 0 {
		t.Errorf("expected 0 UIMessages, got %d", len(pr.Response.UIMessages.Messages))
	}
}

func TestResponse_AddEntity_BasicFields(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	if len(pr.Response.Entities.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(pr.Response.Entities.Entities))
	}
	e := pr.Response.Entities.Entities[0]
	if e.Type != EntityIPv4Address {
		t.Errorf("Type = %q, want %q", e.Type, EntityIPv4Address)
	}
	if e.Value != "1.2.3.4" {
		t.Errorf("Value = %q, want 1.2.3.4", e.Value)
	}
	if e.Weight != 100 {
		t.Errorf("Weight = %d, want 100 (default)", e.Weight)
	}
	if e.IconURL != "" {
		t.Errorf("IconURL = %q, want empty by default", e.IconURL)
	}
}

func TestResponse_AddEntity_MultipleEntities(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.1.1.1")
	r.AddEntity(EntityIPv4Address, "2.2.2.2")
	r.AddEntity(EntityPerson, "SomeActor")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	if len(pr.Response.Entities.Entities) != 3 {
		t.Errorf("expected 3 entities, got %d", len(pr.Response.Entities.Entities))
	}
}

func TestEntityBuilder_SetWeight(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").SetWeight(42)

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	if pr.Response.Entities.Entities[0].Weight != 42 {
		t.Errorf("Weight = %d, want 42", pr.Response.Entities.Entities[0].Weight)
	}
}

func TestEntityBuilder_AddProperty(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").
		AddProperty("classification", "Classification", MatchingRuleLoose, "malicious")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	e := pr.Response.Entities.Entities[0]
	if e.AdditionalFields == nil || len(e.AdditionalFields.Fields) != 1 {
		t.Fatal("expected 1 AdditionalField")
	}
	f := e.AdditionalFields.Fields[0]
	if f.Name != "classification" {
		t.Errorf("Field Name = %q, want classification", f.Name)
	}
	if f.DisplayName != "Classification" {
		t.Errorf("Field DisplayName = %q, want Classification", f.DisplayName)
	}
	if f.MatchingRule != MatchingRuleLoose {
		t.Errorf("Field MatchingRule = %q, want %q", f.MatchingRule, MatchingRuleLoose)
	}
	if f.Value != "malicious" {
		t.Errorf("Field Value = %q, want malicious", f.Value)
	}
}

func TestEntityBuilder_AddDisplayInfo(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").
		AddDisplayInfo("GreyNoise Info", "<b>hello</b>")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	e := pr.Response.Entities.Entities[0]
	if e.DisplayInformation == nil || len(e.DisplayInformation.Labels) != 1 {
		t.Fatal("expected 1 DisplayInformation label")
	}
	lbl := e.DisplayInformation.Labels[0]
	if lbl.Name != "GreyNoise Info" {
		t.Errorf("Label Name = %q, want GreyNoise Info", lbl.Name)
	}
	if !strings.Contains(lbl.Content, "hello") {
		t.Errorf("Label content %q does not contain 'hello'", lbl.Content)
	}
}

func TestEntityBuilder_AddDisplayInfo_EscapesHTML(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").AddDisplayInfo("Info", "<b>hello</b>")
	data, _ := r.ToXML()
	s := string(data)
	if strings.Contains(s, "CDATA") {
		t.Error("display information should not use CDATA")
	}
	if !strings.Contains(s, "&lt;b&gt;hello&lt;/b&gt;") {
		t.Errorf("expected escaped HTML in DisplayInformation label, got:\n%s", s)
	}
}

func TestEntityBuilder_AddDisplayInfo_DoesNotEscapeQuotes(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").AddDisplayInfo("Info", `<a href="https://example.test">open</a>`)
	data, _ := r.ToXML()
	s := string(data)
	if strings.Contains(s, "&#34;") || strings.Contains(s, "&quot;") {
		t.Errorf("display information should keep quotes like Python maltego-trx output, got:\n%s", s)
	}
	if !strings.Contains(s, `href="https://example.test"`) {
		t.Errorf("expected literal quoted href in escaped display content, got:\n%s", s)
	}
}

func TestEntityBuilder_AddDisplayInfo_EscapesAmpersand(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").AddDisplayInfo("Info", `a & b`)
	data, _ := r.ToXML()
	if !strings.Contains(string(data), "a &amp; b") {
		t.Errorf("expected escaped ampersand, got:\n%s", string(data))
	}
}

func TestEntityBuilder_NoProperties_OmitsAdditionalFields(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4")
	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)
	if pr.Response.Entities.Entities[0].AdditionalFields != nil {
		t.Error("AdditionalFields should be omitted when no properties added")
	}
}

func TestEntityBuilder_Chaining(t *testing.T) {
	r := NewResponse()
	r.AddEntity(EntityIPv4Address, "1.2.3.4").
		SetWeight(80).
		AddProperty("a", "A", MatchingRuleLoose, "1").
		AddProperty("b", "B", MatchingRuleStrict, "2").
		AddDisplayInfo("Title", "content")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)
	e := pr.Response.Entities.Entities[0]

	if e.Weight != 80 {
		t.Errorf("Weight = %d, want 80", e.Weight)
	}
	if len(e.AdditionalFields.Fields) != 2 {
		t.Errorf("expected 2 properties, got %d", len(e.AdditionalFields.Fields))
	}
}

// --- UIMessage tests ---

func TestResponse_Inform(t *testing.T) {
	r := NewResponse()
	r.Inform("all good")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	if len(pr.Response.UIMessages.Messages) != 1 {
		t.Fatalf("expected 1 UIMessage, got %d", len(pr.Response.UIMessages.Messages))
	}
	m := pr.Response.UIMessages.Messages[0]
	if m.MessageType != MsgTypeInform {
		t.Errorf("MessageType = %q, want %q", m.MessageType, MsgTypeInform)
	}
	if m.Value != "all good" {
		t.Errorf("Message = %q, want 'all good'", m.Value)
	}
}

func TestResponse_FatalError(t *testing.T) {
	r := NewResponse()
	r.FatalError("something broke")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	m := pr.Response.UIMessages.Messages[0]
	if m.MessageType != MsgTypeFatalError {
		t.Errorf("MessageType = %q, want %q", m.MessageType, MsgTypeFatalError)
	}
	if m.Value != "something broke" {
		t.Errorf("Message = %q, want 'something broke'", m.Value)
	}
}

func TestResponse_MultipleUIMessages(t *testing.T) {
	r := NewResponse()
	r.Inform("msg1")
	r.FatalError("msg2")

	data, _ := r.ToXML()
	pr := parseXMLResponse(t, data)

	if len(pr.Response.UIMessages.Messages) != 2 {
		t.Errorf("expected 2 UIMessages, got %d", len(pr.Response.UIMessages.Messages))
	}
}

// --- ErrorResponse helper ---

func TestErrorResponse(t *testing.T) {
	data, err := ErrorResponse("boom")
	if err != nil {
		t.Fatalf("ErrorResponse error: %v", err)
	}

	pr := parseXMLResponse(t, data)

	if len(pr.Response.UIMessages.Messages) != 1 {
		t.Fatalf("expected 1 UIMessage, got %d", len(pr.Response.UIMessages.Messages))
	}
	if pr.Response.UIMessages.Messages[0].MessageType != MsgTypeFatalError {
		t.Error("expected FatalError message type")
	}
	if len(pr.Response.Entities.Entities) != 0 {
		t.Errorf("expected 0 entities in error response, got %d", len(pr.Response.Entities.Entities))
	}
}

// --- XML structure ---

func TestToXML_OmitsXMLHeaderLikePythonTRX(t *testing.T) {
	r := NewResponse()
	data, _ := r.ToXML()
	if strings.HasPrefix(string(data), "<?xml") {
		t.Error("response should omit XML header to match Python maltego-trx output")
	}
}

func TestToXML_ContainsRequiredTags(t *testing.T) {
	r := NewResponse()
	data, _ := r.ToXML()
	s := string(data)

	for _, tag := range []string{"MaltegoMessage", "MaltegoTransformResponseMessage", "Entities", "UIMessages"} {
		if !strings.Contains(s, tag) {
			t.Errorf("XML missing required tag <%s>", tag)
		}
	}
}
