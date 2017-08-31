package main

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

func genHTML(doc document) ([]byte, error) {
	var buf bytes.Buffer
	write := func(s string) {
		buf.WriteString(s)
	}

	write(`<!DOCTYPE html><html>`)
	if doc.title != "" {
		write(`<head><title>` + html.EscapeString(doc.title) + `</title></head>`)
	}
	write(`<body>`)
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
		default:
			return nil, fmt.Errorf("unhandled document part: %#v", p)
		}
	}
	write(`</body></html>`)

	return buf.Bytes(), nil
}
