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
<style>
 body{
  background-color: #D7EEEF;
  text-align: left;
  max-width:800px;
  margin-left: auto;
  margin-right: auto;
 }
</style>`)
	if doc.title != "" {
		write(`<title>` + escapeHTML(doc.title) + `</title>`)
	}
	write(`</head><body>`)
	for _, part := range doc.parts {
		switch p := part.(type) {
		case docText:
			lines := strings.Split(string(p), "\n")
			for i := range lines {
				lines[i] = escapeHTML(lines[i])
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
		case docTitle:
			write("<h1>" + string(p) + "</h1>")
		case docCaption:
			write("<h2>" + string(p) + "</h2>")
		case docSubCaption:
			write("<h3>" + string(p) + "</h3>")
		case docSubSubCaption:
			write("<h4>" + string(p) + "</h4>")
		case docLink:
			write(fmt.Sprintf(`<a href="#%d">%s</a>`, p.id, escapeHTML(p.text)))
		case docLinkTarget:
			write(fmt.Sprintf(`<a id="%d"/>`, int(p)))
		case externalDocLink:
			write(fmt.Sprintf(`<a href="%s">%s</a>`, p.url, p.text))
		case stylizedDocText:
			if p.bold {
				write("<b>")
			}
			if p.italic {
				write("<i>")
			}
			write(escapeHTML(p.text))
			if p.italic {
				write("</i>")
			}
			if p.bold {
				write("</b>")
			}
		default:
			return nil, fmt.Errorf("error generating HTML: unhandled document part: %T", p)
		}
	}
	write(`</body></html>`)

	return buf.Bytes(), nil
}

func escapeHTML(s string) string {
	s = strings.Replace(s, "\t", "    ", -1)
	s = html.EscapeString(s)
	s = strings.Replace(s, "  ", "&nbsp;&nbsp;", -1)
	s = strings.Replace(s, "&nbsp; ", "&nbsp;&nbsp;", -1)
	return s
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
