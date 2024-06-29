package utils

import (
	"bytes"
	"encoding/xml"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
)

// GBKXmlUnmarshal 解析`<?xml version="1.0" encoding="GB2312"?>`开头的XML
func GBKXmlUnmarshal(xmlString string, v any) error {
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlString)))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return transform.NewReader(input, simplifiedchinese.GBK.NewEncoder()), nil
	}
	return decoder.Decode(v)
}

// GBKXmlGetRoot 函数尝试解析XML的根元素名称
func GBKXmlGetRoot(xmlString string) (string, error) {
	var root struct {
		XMLName xml.Name
	}

	err := GBKXmlUnmarshal(xmlString, &root)
	if err != nil {
		return "", err
	}

	return root.XMLName.Local, nil
}
