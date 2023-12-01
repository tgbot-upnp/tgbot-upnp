package upnp

import (
	"encoding/xml"
	"github.com/gotd/td/tg"
)

type didLLite struct {
	XMLName      xml.Name     `xml:"DIDL-Lite"`
	SchemaDIDL   string       `xml:"xmlns,attr"`
	DC           string       `xml:"xmlns:dc,attr"`
	Sec          string       `xml:"xmlns:sec,attr"`
	SchemaUPNP   string       `xml:"xmlns:upnp,attr"`
	DIDLLiteItem didLLiteItem `xml:"item"`
}

type didLLiteItem struct {
	SecCaptionInfo   *secCaptionInfo   `xml:"sec:CaptionInfo,omitempty"`
	SecCaptionInfoEx *secCaptionInfoEx `xml:"sec:CaptionInfoEx,omitempty"`
	XMLName          xml.Name          `xml:"item"`
	DCtitle          string            `xml:"dc:title"`
	UPNPClass        string            `xml:"upnp:class"`
	ID               string            `xml:"id,attr"`
	ParentID         string            `xml:"parentID,attr"`
	Restricted       string            `xml:"restricted,attr"`
}
type secCaptionInfo struct {
	XMLName xml.Name `xml:"sec:CaptionInfo"`
	Type    string   `xml:"sec:type,attr"`
	Value   string   `xml:",chardata"`
}

type secCaptionInfoEx struct {
	XMLName xml.Name `xml:"sec:CaptionInfoEx"`
	Type    string   `xml:"sec:type,attr"`
	Value   string   `xml:",chardata"`
}

func GetMetaData(message, tgVideoID string, media *tg.Document) string {
	mediaTitle := message
	if mediaTitle == "" {
		if len(media.MapAttributes().AsDocumentAttributeFilename()) > 0 {
			mediaTitle = media.MapAttributes().AsDocumentAttributeFilename()[0].FileName
		} else {
			mediaTitle = tgVideoID
		}
	}
	metaData, _ := xml.Marshal(didLLite{
		XMLName:    xml.Name{},
		SchemaDIDL: "urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/",
		DC:         "http://purl.org/dc/elements/1.1/",
		Sec:        "http://www.sec.co.kr/",
		SchemaUPNP: "urn:schemas-upnp-org:metadata-1-0/upnp/",
		DIDLLiteItem: didLLiteItem{
			XMLName:    xml.Name{},
			ID:         "1",
			ParentID:   "0",
			Restricted: "1",
			UPNPClass:  "object.item.videoItem.movie",
			DCtitle:    mediaTitle,
		},
	})
	return string(metaData)
}
