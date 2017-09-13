package main

import (
	"bytes"
	"fmt"
	"image/png"
	"strings"
)

func genRTF(doc document) ([]byte, error) {
	hexChars := []byte("0123456789abcdef")
	const maxImageW = 780

	var buf bytes.Buffer
	write := func(s string) {
		buf.WriteString(s)
	}
	writeCaption := func(s string, size string) {
		if size != "" {
			size = `\fs` + size
		}
		write(`\b` + size + ` `)
		write(strings.Replace(s, "\n", `\line `, -1))
		write(`\b0\fs22\line `)
	}

	write(`{\rtf1\ansi\deff0{\fonttbl{\f0\fnil\fcharset0 Calibri;}}`)
	for _, part := range doc.parts {
		switch p := part.(type) {
		case docText:
			write(strings.Replace(string(p), "\n", `\line `, -1))
		case docImage:
			img, err := findImage(p.name)
			if err != nil {
				return nil, fmt.Errorf("error generating RTF image '%s': %s", p.name, err.Error())
			}
			w, h := img.Bounds().Dx(), img.Bounds().Dy()
			destW, destH := w, h
			if destW > maxImageW {
				scale := maxImageW / float64(destW)
				destW = maxImageW
				destH = int(float64(destH)*scale + 0.5)
			}
			size := fmt.Sprintf(
				`\picw%d\pich%d\picwgoal%d\pichgoal%d `,
				toTwips(w),
				toTwips(h),
				toTwips(destW),
				toTwips(destH),
			)
			write(`{\*\shppict{\pict\pngblip` + size)
			var imgBuf bytes.Buffer
			err = png.Encode(&imgBuf, img)
			if err != nil {
				return nil, fmt.Errorf("error encoding RTF png '%s': %s", p.name, err.Error())
			}
			hex := make([]byte, imgBuf.Len()*2)
			for i, b := range imgBuf.Bytes() {
				hex[i*2] = hexChars[b&0xF0>>4]
				hex[i*2+1] = hexChars[b&0x0F]
			}
			buf.Write(hex)
			write("\n}}")
		case docTitle:
			writeCaption(string(p), "45")
		case docCaption:
			writeCaption(string(p), "40")
		case docSubCaption:
			writeCaption(string(p), "34")
		case docSubSubCaption:
			writeCaption(string(p), "")
		case docLink:
			write(strings.Replace(string(p.text), "\n", `\line `, -1))
		case docLinkTarget:
			// NOTE there are no links in RTF
		case externalDocLink:
			write(fmt.Sprintf(`{\field{\*\fldinst HYPERLINK "%s"}{\fldrslt %s}}`, p.url, p.text))
		case stylizedDocText:
			if p.bold {
				write(`\b `)
			}
			if p.italic {
				write(`\i `)
			}
			write(strings.Replace(p.text, "\n", `\line `, -1))
			if p.italic {
				write(`\i0 `)
			}
			if p.bold {
				write(`\b0 `)
			}
		default:
			return nil, fmt.Errorf("error generating RTF: unhandled document part: %T", p)
		}
	}
	write(`}`)

	return buf.Bytes(), nil
}

func toTwips(x int) int {
	// see https://stackoverflow.com/questions/1490734/programmatically-adding-images-to-rtf-document
	return x * 1440 / 96
}
