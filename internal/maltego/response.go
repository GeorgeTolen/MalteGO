package maltego

import (
	"encoding/xml"
	"strings"
)

// --- Response builder ---

type Response struct {
	entities   []respEntity
	uiMessages []respUIMessage
}

type respEntity struct {
	Type               string
	Value              string
	Weight             int
	LinkLabel          string
	DisplayInformation []Label
	Properties         []Field
	Overlays           []Overlay
	IconURL            string
}

type Label struct {
	Name    string
	Type    string
	Content string
}

type Field struct {
	Name         string
	DisplayName  string
	MatchingRule string
	Value        string
}

type Overlay struct {
	PropertyName string
	Position     string
	Type         string
}

type respUIMessage struct {
	Type    string
	Message string
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) AddEntity(entityType, value string) *EntityBuilder {
	e := respEntity{
		Type:   entityType,
		Value:  value,
		Weight: 100,
	}
	r.entities = append(r.entities, e)
	return &EntityBuilder{resp: r, idx: len(r.entities) - 1}
}

func (r *Response) AddUIMessage(msg, msgType string) {
	r.uiMessages = append(r.uiMessages, respUIMessage{Type: msgType, Message: msg})
}

func (r *Response) Inform(msg string) {
	r.AddUIMessage(msg, MsgTypeInform)
}

func (r *Response) FatalError(msg string) {
	r.AddUIMessage(msg, MsgTypeFatalError)
}

// EntityBuilder allows chaining entity properties.
type EntityBuilder struct {
	resp *Response
	idx  int
}

func (b *EntityBuilder) SetWeight(w int) *EntityBuilder {
	b.resp.entities[b.idx].Weight = w
	return b
}

func (b *EntityBuilder) SetIconURL(url string) *EntityBuilder {
	b.resp.entities[b.idx].IconURL = url
	return b
}

func (b *EntityBuilder) SetLinkLabel(label string) *EntityBuilder {
	b.resp.entities[b.idx].LinkLabel = label
	return b
}

func (b *EntityBuilder) AddDisplayInfo(name, content string) *EntityBuilder {
	b.resp.entities[b.idx].DisplayInformation = append(
		b.resp.entities[b.idx].DisplayInformation,
		Label{Name: name, Type: "text/html", Content: content},
	)
	return b
}

func (b *EntityBuilder) AddProperty(name, displayName, matchingRule, value string) *EntityBuilder {
	b.resp.entities[b.idx].Properties = append(
		b.resp.entities[b.idx].Properties,
		Field{Name: name, DisplayName: displayName, MatchingRule: matchingRule, Value: value},
	)
	return b
}

func (b *EntityBuilder) AddOverlay(propertyName, position, overlayType string) *EntityBuilder {
	b.resp.entities[b.idx].Overlays = append(
		b.resp.entities[b.idx].Overlays,
		Overlay{PropertyName: propertyName, Position: position, Type: overlayType},
	)
	return b
}

// --- XML serialisation ---

type xmlResponse struct {
	XMLName  xml.Name          `xml:"MaltegoMessage"`
	Response xmlTransformResp  `xml:"MaltegoTransformResponseMessage"`
}

type xmlTransformResp struct {
	Entities   xmlRespEntities  `xml:"Entities"`
	UIMessages xmlUIMessages    `xml:"UIMessages"`
}

type xmlRespEntities struct {
	Entities []xmlRespEntity `xml:"Entity"`
}

type xmlRespEntity struct {
	Type               string              `xml:"Type,attr"`
	Value              xmlText             `xml:"Value"`
	Weight             int                 `xml:"Weight"`
	LinkLabel          string              `xml:"LinkLabel,omitempty"`
	DisplayInformation *xmlDisplayInfo     `xml:"DisplayInformation,omitempty"`
	AdditionalFields   *xmlRespFields      `xml:"AdditionalFields,omitempty"`
	Overlays           *xmlOverlays        `xml:"Overlays,omitempty"`
	IconURL            string              `xml:"IconURL,omitempty"`
}

type xmlDisplayInfo struct {
	Labels []xmlLabel `xml:"Label"`
}

type xmlLabel struct {
	Name    string `xml:"Name,attr"`
	Type    string `xml:"Type,attr"`
	Content string `xml:",innerxml"`
}

type xmlText struct {
	Content string `xml:",innerxml"`
}

type xmlRespFields struct {
	Fields []xmlRespField `xml:"Field"`
}

type xmlRespField struct {
	DisplayName  string `xml:"DisplayName,attr"`
	MatchingRule string `xml:"MatchingRule,attr,omitempty"`
	Name         string `xml:"Name,attr"`
	Value        string `xml:",innerxml"`
}

type xmlOverlays struct {
	Overlays []xmlOverlay `xml:"Overlay"`
}

type xmlOverlay struct {
	Position     string `xml:"position,attr"`
	PropertyName string `xml:"propertyName,attr"`
	Type         string `xml:"type,attr"`
}

type xmlUIMessages struct {
	Messages []xmlUIMessage `xml:"UIMessage"`
}

type xmlUIMessage struct {
	MessageType string `xml:"MessageType,attr"`
	Value       string `xml:",innerxml"`
}

func (r *Response) ToXML() ([]byte, error) {
	out := xmlResponse{
		Response: xmlTransformResp{
			Entities:   xmlRespEntities{},
			UIMessages: xmlUIMessages{},
		},
	}

	for _, e := range r.entities {
		xe := xmlRespEntity{
			Type:      e.Type,
			Value:     xmlText{Content: escapeTextContent(e.Value)},
			Weight:    e.Weight,
			LinkLabel: e.LinkLabel,
			IconURL:   e.IconURL,
		}

		if len(e.DisplayInformation) > 0 {
			di := &xmlDisplayInfo{}
			for _, l := range e.DisplayInformation {
				di.Labels = append(di.Labels, xmlLabel{
					Name:    l.Name,
					Type:    l.Type,
					Content: escapeDisplayContent(l.Content),
				})
			}
			xe.DisplayInformation = di
		}

		if len(e.Properties) > 0 {
			rf := &xmlRespFields{}
			for _, p := range e.Properties {
				rf.Fields = append(rf.Fields, xmlRespField{
					Name:         p.Name,
					DisplayName:  p.DisplayName,
					MatchingRule: p.MatchingRule,
					Value:        escapeTextContent(p.Value),
				})
			}
			xe.AdditionalFields = rf
		}

		if len(e.Overlays) > 0 {
			overlays := &xmlOverlays{}
			for _, o := range e.Overlays {
				overlays.Overlays = append(overlays.Overlays, xmlOverlay{
					Position:     o.Position,
					PropertyName: o.PropertyName,
					Type:         o.Type,
				})
			}
			xe.Overlays = overlays
		}

		out.Response.Entities.Entities = append(out.Response.Entities.Entities, xe)
	}

	for _, m := range r.uiMessages {
		out.Response.UIMessages.Messages = append(out.Response.UIMessages.Messages, xmlUIMessage{
			MessageType: m.Type,
			Value:       escapeTextContent(m.Message),
		})
	}

	data, err := xml.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func escapeDisplayContent(s string) string {
	return escapeTextContent(s)
}

func escapeTextContent(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(s)
}

func ErrorResponse(msg string) ([]byte, error) {
	r := NewResponse()
	r.FatalError(msg)
	return r.ToXML()
}
