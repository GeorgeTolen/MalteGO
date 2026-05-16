package maltego

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// --- XML input structures ---

type xmlMessage struct {
	XMLName xml.Name   `xml:"MaltegoMessage"`
	Request xmlRequest `xml:"MaltegoTransformRequestMessage"`
}

type xmlRequest struct {
	Entities        xmlEntities        `xml:"Entities"`
	TransformFields xmlTransformFields `xml:"TransformFields"`
	Limits          xmlLimits          `xml:"Limits"`
}

type xmlEntities struct {
	Entity xmlEntity `xml:"Entity"`
}

type xmlEntity struct {
	Type             string            `xml:"Type,attr"`
	Value            string            `xml:"Value"`
	Weight           string            `xml:"Weight"`
	AdditionalFields xmlAdditionalFields `xml:"AdditionalFields"`
}

type xmlAdditionalFields struct {
	Fields []xmlField `xml:"Field"`
}

type xmlTransformFields struct {
	Fields []xmlField `xml:"Field"`
}

type xmlField struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",chardata"`
}

type xmlLimits struct {
	SoftLimit string `xml:"SoftLimit,attr"`
	HardLimit string `xml:"HardLimit,attr"`
}

// --- Parsed request ---

type Request struct {
	Value      string
	EntityType string
	Properties map[string]string
	Settings   map[string]string
	SoftLimit  int
	HardLimit  int
}

func ParseRequest(data []byte) (*Request, error) {
	var msg xmlMessage
	if err := xml.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("xml parse: %w", err)
	}

	props := make(map[string]string)
	for _, f := range msg.Request.Entities.Entity.AdditionalFields.Fields {
		props[f.Name] = f.Value
	}

	settings := make(map[string]string)
	for _, f := range msg.Request.TransformFields.Fields {
		settings[f.Name] = f.Value
	}

	softLimit, _ := strconv.Atoi(msg.Request.Limits.SoftLimit)
	hardLimit, _ := strconv.Atoi(msg.Request.Limits.HardLimit)
	if softLimit == 0 {
		softLimit = 12
	}
	if hardLimit == 0 {
		hardLimit = 12
	}

	return &Request{
		Value:      msg.Request.Entities.Entity.Value,
		EntityType: msg.Request.Entities.Entity.Type,
		Properties: props,
		Settings:   settings,
		SoftLimit:  softLimit,
		HardLimit:  hardLimit,
	}, nil
}

func (r *Request) APIKey(fallback string) string {
	if k := r.Settings["greynoise.api.key"]; k != "" {
		return k
	}
	return fallback
}
