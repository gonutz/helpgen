package main

import (
	"fmt"
	"image/color"
	"strings"

	rtf "github.com/therox/rtf-doc"
)

func genRTF(doc document) ([]byte, error) {
	rtfDoc := rtf.NewDocument()
	rtfDoc.SetFormat("A4")
	rtfDoc.SetOrientation("portrait")
	fonts := rtfDoc.NewFontTable()
	fonts.AddFont("swiss", 0, 0, "Arial", "arial")
	colors := rtfDoc.NewColorTable()
	colors.AddColor(color.RGBA{R: 0, G: 0, B: 0, A: 255}, "black")

	for _, part := range doc.parts {
		switch p := part.(type) {
		case docText:
			para := rtfDoc.AddParagraph()
			lines := strings.Split(string(p), "\n")
			for _, line := range lines {
				para.AddText(line, 12, "arial", "")
				para.AddNewLine()
			}
		case docImage:
		case largerDocText:
		default:
			return nil, fmt.Errorf("unhandled document part: %#v", p)
		}
	}

	return rtfDoc.Export(), nil
}
