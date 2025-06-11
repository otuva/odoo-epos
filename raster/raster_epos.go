package raster

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strconv"
)

// NewRasterImage 从XML数据中解析并返回RasterImage对象
func NewRasterImageFromXML(payload []byte) (*RasterImage, error) {
	envelope := &Envelope{}
	if err := xml.Unmarshal(payload, envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SOAP envelope: %w", err)
	}
	return envelope.Body.EposPrint.Image, nil
}

// Envelope 表示SOAP信封结构
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

// Body 表示SOAP消息体
type Body struct {
	XMLName   xml.Name  `xml:"Body"`
	EposPrint EposPrint `xml:"epos-print"`
}

// EposPrint 表示打印机指令容器
type EposPrint struct {
	XMLName xml.Name     `xml:"epos-print"`
	Xmlns   string       `xml:"xmlns,attr"`
	Image   *RasterImage `xml:"image"`
}

func (img *RasterImage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias RasterImage
	aux := &struct {
		Width   string `xml:"width,attr"`
		Height  string `xml:"height,attr"`
		Content string `xml:",chardata"`
		*Alias
	}{
		Alias: (*Alias)(img),
	}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}

	var err error
	img.Width, err = strconv.Atoi(aux.Width)
	if err != nil {
		return err
	}

	img.Height, err = strconv.Atoi(aux.Height)
	if err != nil {
		return err
	}

	img.Content, err = base64.StdEncoding.DecodeString(aux.Content)
	if err != nil {
		return err
	}

	return nil
}
