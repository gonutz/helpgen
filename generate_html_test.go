package main

import (
	"strings"
	"testing"
)

func TestMultipleSpacesAreKept(t *testing.T) {
	checkHTMLbody(t, "two&nbsp;&nbsp;spaces", "", docText("two  spaces"))
	checkHTMLbody(t, "three&nbsp;&nbsp;&nbsp;spaces", "", docText("three   spaces"))
	checkHTMLbody(t, "4&nbsp;&nbsp;&nbsp;&nbsp;spaces", "", docText("4    spaces"))
}

func TestTabsAreReplacedByFourSpaces(t *testing.T) {
	checkHTMLbody(t, "a&nbsp;&nbsp;&nbsp;&nbsp;tab", "", docText("a\ttab"))
}

func TestTrademarkRIsSuperscripted(t *testing.T) {
	checkHTMLbody(t, "<sup>®</sup>", "", docText("®"))
}

func checkHTMLbody(t *testing.T, want string, docTitle string, docParts ...docPart) {
	doc := document{title: docTitle, parts: docParts}
	output, err := genHTML(doc)
	if err != nil {
		t.Fatal("got error:", err)
	}
	html := string(output)
	start := strings.Index(html, "<body>") + len("<body>")
	end := strings.LastIndex(html, "</body>")
	body := html[start:end]
	if body != want {
		t.Errorf("HTML body differs, want\n'%s'\nbut have\n'%s'", want, body)
	}
}
