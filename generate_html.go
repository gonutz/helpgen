package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"image"
	"image/png"
	"strings"
)

func genHTML(doc document) ([]byte, error) {
	var buf bytes.Buffer
	write := func(s string) {
		buf.WriteString(s)
	}

	write(`<!DOCTYPE html><meta charset="UTF-8"><html><head>
<style>body{
background-color: #D7EEEF;
text-align: center;
max-width:800px;
margin: 0 auto !important;
float: none !important;
}</style>`)
	if doc.title != "" {
		write(`<title>` + html.EscapeString(doc.title) + `</title>`)
	}
	write(`</head><body>`)
	for _, part := range doc.parts {
		switch p := part.(type) {
		case docText:
			lines := strings.Split(string(p), "\n")
			for i := range lines {
				lines[i] = html.EscapeString(lines[i])
			}
			write(strings.Join(lines, "<br>"))
		case docImage:
			img, err := findImage(p.name)
			if err != nil {
				return nil, fmt.Errorf("error generating HTML image '%s': %s", p.name, err.Error())
			}
			tag, err := imageTag(img)
			if err != nil {
				return nil, fmt.Errorf("error generating HTML image tag for '%s': %s", p.name, err.Error())
			}
			write(tag)
		case largerDocText:
			// map scale heading
			//       2     3
			//       3     2
			//       4     1
			tag := fmt.Sprintf("h%d", 5-p.scale)
			write("<" + tag + ">")
			write(p.text)
			write("</" + tag + ">")
		case docLink:
			write(fmt.Sprintf(`<a href="#%s">%s</a>`, p.id, html.EscapeString(p.text)))
		case docLinkTarget:
			write(fmt.Sprintf(`<a id="%s"/>`, p.id))
		default:
			return nil, fmt.Errorf("error generating HTML: unhandled document part: %#v", p)
		}
	}
	write(`</body></html>`)

	return buf.Bytes(), nil
}

func imageTag(img image.Image) (string, error) {
	var buf bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &buf)
	err := png.Encode(e, img)
	if err != nil {
		return "", errors.New("cannot encode image as PNG: " + err.Error())
	}
	err = e.Close()
	if err != nil {
		return "", errors.New("cannot encode image as Base64: " + err.Error())
	}
	return `<img src="data:image/png;base64,` + string(buf.Bytes()) + `">`, nil
}
