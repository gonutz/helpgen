package main

import (
	"reflect"
	"testing"
)

func TestDocumentTitle(t *testing.T) {
	checkParse(
		t,
		"=====\nTitle\n=====",
		"Title",
		docTitle("Title"),
	)
}

func TestTitleFollowedByCaption(t *testing.T) {
	checkParse(
		t,
		`===
Title
===
Caption
=======`,
		"Title",
		docTitle("Title"),
		docCaption("Caption"),
	)
}

func TestThereCanOnlyBeOneTitle(t *testing.T) {
	checkParseError(
		t,
		`===
Title
===
===
2nd
===`,
		"title redefined in line 5, first definition in line 2, there can only be one title",
	)
}

func TestTitlesAreNeverStylzed(t *testing.T) {
	const title = `*no bold* /nor italic/ [nor references]`
	checkParse(
		t,
		"===\n"+title+"\n===",
		title,
		docTitle(title),
	)
}

func TestSingleLineOfText(t *testing.T) {
	const text = "This is plain text."
	checkParse(t, text, "", docText(text))
}

func TestTwoLinesOfText(t *testing.T) {
	const text = "Line one\nLine two"
	checkParse(t, text, "", docText(text))
}

func TestWindowsLineBreaksAreReplacedWithUnix(t *testing.T) {
	checkParse(t, "Line one\r\nLine two", "", docText("Line one\nLine two"))
}

func TestOldMaxLineBreaksAreReplacedWithUnix(t *testing.T) {
	checkParse(t, "Line one\rLine two", "", docText("Line one\nLine two"))
}

func TestVariablesCanOnlyBeDefinedOnce(t *testing.T) {
	checkParseError(
		t,
		`[\var=text]
[\var=text]`,
		`variable 'var' redefined in line 2, first definition was in line 1, each variable can only be defined once`,
	)
}

func TestVariablesAreReplacedInText(t *testing.T) {
	checkParse(t, `[\var=text]
before [var] after`, "", docText("before text after"))
}

func TestVariablesCanBeDefinedAfterUsage(t *testing.T) {
	checkParse(t, `before [var] after
[\var=text]`, "", docText("before text after"))
}

func TestBoldText(t *testing.T) {
	checkParse(t, "*bold*", "", bold("bold"))
}

func TestNonBoldWithStyleCharacter(t *testing.T) {
	checkParse(t, "not* bold*", "", docText("not* bold*"))
}

func TestItalicText(t *testing.T) {
	checkParse(t, "/italic/", "", italic("italic"))
	checkParse(t, "/File/->/Exit/", "", italic("File"), docText("->"), italic("Exit"))
}

func TestBoldItalicText(t *testing.T) {
	checkParse(t, "*/both/*", "", boldItalic("both"))
	checkParse(t, "/*both*/", "", boldItalic("both"))
}

func TestBoldTextBeforeDot(t *testing.T) {
	checkParse(t, "ends in *bold*.", "", docText("ends in "), bold("bold"), docText("."))
}

func TestStylesCannotOverlap(t *testing.T) {
	checkParse(t, "*bold /both* italic/", "", bold("bold /both"), docText(" italic/"))
}

func TestStylizedTextCanContainVariables(t *testing.T) {
	checkParse(t, `[\var=some text]
*here is [var]*`, "", bold("here is some text"))
	checkParse(t, `[\var=some text]
*here is [var] and again [var]*`, "", bold("here is some text and again some text"))
}

func TestVariablesOnlyContainVerbatimText(t *testing.T) {
	checkParse(t, `[\var=*not bold*]
[var]`, "", docText("*not bold*"))
}

func TestCaptionsCanHaveVariables(t *testing.T) {
	checkParse(t, `===
[title]
===
[\title=abc]`, "abc", docTitle("abc"))
}

func TestStylesAreNotNested(t *testing.T) {
	checkParse(
		t,
		"*bold /NOT italic!/ bold again*",
		"",
		bold("bold /NOT italic!/ bold again"),
	)
}

func TestParseStyles(t *testing.T) {
	checkParse(
		t,
		`*bold* nothing /italic/`,
		"",
		bold("bold"),
		docText(" nothing "),
		italic("italic"),
	)
}

func TestEsacpedSpecialCharacter(t *testing.T) {
	checkParse(t, "[*]", "", docText("*"))
	checkParse(t, "[[]", "", docText("["))
	checkParse(t, "[/]", "", docText("/"))
	checkParse(t, "[=]", "", docText("="))
	checkParse(t, "[-]", "", docText("-"))
	checkParse(t, "[.]", "", docText("."))
}

func TestImageRefsHaveImageExtension(t *testing.T) {
	checkParse(t, "[image.png]", "", docImage{name: "image.png"})
	checkParse(t, "[image.JPG]", "", docImage{name: "image.JPG"})
	checkParse(t, "[image.jPeg]", "", docImage{name: "image.jPeg"})
	checkParse(t, "[image.BMP]", "", docImage{name: "image.BMP"})
	checkParse(t, "[image.gif]", "", docImage{name: "image.gif"})
	checkParse(t, "[.png]", "", docImage{name: ".png"})
}

func TestOnlyKnownImageExtensionsBecomeImages(t *testing.T) {
	checkParseError(t, "[no-image.txt]", "unknown link target 'no-image.txt' in line 1")
}

func TestCaptions(t *testing.T) {
	checkParse(
		t,
		`
Chapter
=======
Subchapter
----------
Subsubchapter
.............`,
		"",
		docText("\n"),
		docCaption("Chapter"),
		docSubCaption("Subchapter"),
		docSubSubCaption("Subsubchapter"),
	)
}

func TestTwoConsecutiveCaptionsAreNotATitle(t *testing.T) {
	checkParse(
		t,
		`
Caption 1
=========
Caption 2
=========
Caption 3
=========`,
		"",
		docText("\n"),
		docCaption("Caption 1"),
		docCaption("Caption 2"),
		docCaption("Caption 3"),
	)
}

func TestSingleChapterCaption(t *testing.T) {
	checkParse(
		t,
		`chap 1
=====`,
		"",
		docCaption("chap 1"),
	)
}

func TestRefsToCaptionsAreResolved(t *testing.T) {
	checkParse(
		t,
		`chap 1
=====
[chap 1]`,
		"",
		docLinkTarget(1),
		docCaption("chap 1"),
		docLink{id: 1, text: "chap 1"},
	)
}

func TestUnknownRefTargetIsError(t *testing.T) {
	checkParseError(t, "[who]", "unknown link target 'who' in line 1")
}

func TestRefsToCaptionsCanHaveDifferentText(t *testing.T) {
	checkParse(
		t,
		`chap 1
------
[this is a link[chap 1]]`,
		"",
		docLinkTarget(1),
		docSubCaption("chap 1"),
		docLink{id: 1, text: "this is a link"},
	)
}

func TestReferencesToWebLinksArePrefixedWith_http_ifNecessary(t *testing.T) {
	checkParse(t, "[www.google.com]", "", externalDocLink{
		url:  "http://www.google.com",
		text: "www.google.com",
	})
	checkParse(t, "[http://www.google.com]", "", externalDocLink{
		url:  "http://www.google.com",
		text: "http://www.google.com",
	})
	checkParse(t, "[https://www.google.com]", "", externalDocLink{
		url:  "https://www.google.com",
		text: "https://www.google.com",
	})
	checkParse(t, "[some link[www.google.com]]", "", externalDocLink{
		url:  "http://www.google.com",
		text: "some link",
	})
}

func TestMailAddressRefsAreLinksWith_mailto_prefixedIfNecessary(t *testing.T) {
	checkParse(t, "[blah@mail.com]", "", externalDocLink{
		url:  "mailto:blah@mail.com",
		text: "blah@mail.com",
	})
	checkParse(t, "[mailto:blah@mail.com]", "", externalDocLink{
		url:  "mailto:blah@mail.com",
		text: "blah@mail.com",
	})
	checkParse(t, "[My Mail[blah@mail.com]]", "", externalDocLink{
		url:  "mailto:blah@mail.com",
		text: "My Mail",
	})
}

func checkParse(t *testing.T, code string, title string, want ...docPart) {
	doc, err := parse([]byte(code))
	if err != nil {
		t.Error("parse error:", err)
		return
	}
	if doc.title != title {
		t.Errorf("wrong title, want %s, got %s", title, doc.title)
	}
	if len(want) != len(doc.parts) {
		t.Errorf("want %d parts, got %d: %#v", len(want), len(doc.parts), doc.parts)
		return
	}
	for i := range want {
		a, b := reflect.TypeOf(want[i]), reflect.TypeOf(doc.parts[i])
		if a != b {
			t.Errorf("part %d differs in type, want %v, got %v", i, a, b)
		}
		if !reflect.DeepEqual(want[i], doc.parts[i]) {
			t.Errorf("part %d differs, want '%v', got '%v'", i, want[i], doc.parts[i])
		}
	}
}

// checkParseError takes the expected error message as a parameter, if you leave
// it empty, the function only checks that there is any error at all, not
// comparing the message
func checkParseError(t *testing.T, code string, wantMsg string) {
	_, err := parse([]byte(code))
	if err == nil {
		t.Error("error expected but was none")
		return
	}
	msg := err.Error()
	if wantMsg != "" && msg != wantMsg {
		t.Errorf("expected error message '%s' but got '%s'", wantMsg, msg)
	}
}

func bold(s string) stylizedDocText {
	return stylizedDocText{
		bold: true,
		text: s,
	}
}

func italic(s string) stylizedDocText {
	return stylizedDocText{
		italic: true,
		text:   s,
	}
}

func boldItalic(s string) stylizedDocText {
	return stylizedDocText{
		bold:   true,
		italic: true,
		text:   s,
	}
}
