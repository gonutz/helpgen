package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func parse(code []byte) (document, error) {
	var p parser
	p.code = code
	p.parse()
	return p.doc, p.err
}

type parser struct {
	code      []byte
	text      string
	cur       int
	col, line int
	doc       document
	err       error
}

func (p *parser) parse() {
	for p.err == nil {
		r := p.next()
		if r == eof {
			p.finishText()
			return
		} else if r == '\\' {
			p.finishText()
			p.parseCommand()
		} else if r == '\r' {
			// skip carriage returns
		} else {
			p.text += string(r)
		}
	}
}

func (p *parser) emit(part docPart) {
	p.doc.parts = append(p.doc.parts, part)
}

func (p *parser) finishText() {
	if len(p.text) > 0 {
		p.emit(docText(p.text))
		p.text = ""
	}
}

func (p *parser) parseCommand() {
	for i := range commands {
		if bytes.HasPrefix(p.code[p.cur:], commands[i].text) {
			p.cur += len(commands[i].text)
			r := p.next()
			if r != '(' {
				p.errorf("'(' expected after %s command", commands[i].text)
				return
			}
			start := p.cur
			for r != eof && r != ')' {
				r = p.next()
			}
			if r == eof {
				p.errorf("')' expected after %s command parameter", commands[i].text)
				return
			}
			param := string(p.code[start : p.cur-1])
			if commands[i].trimWhiteSpace {
				param = strings.TrimSpace(param)
			}

			switch commands[i].id {
			case cmdTitle:
				p.doc.title = param
				p.eatWhiteSpace()
			case cmdImage:
				p.emit(docImage{name: param})
			case cmdMagnify2:
				p.emit(largerDocText{text: param, scale: 2})
			case cmdMagnify3:
				p.emit(largerDocText{text: param, scale: 3})
			case cmdMagnify4:
				p.emit(largerDocText{text: param, scale: 4})
			default:
				p.errorf("unhandled command: '%s'", commands[i].text)
			}
			break
		}
	}
}

const eof rune = 0

func (p *parser) eatWhiteSpace() {
	for unicode.IsSpace(p.peek()) {
		p.next()
	}
}

func (p *parser) peek() rune {
	if p.cur >= len(p.code) {
		return eof
	}
	r, _ := utf8.DecodeRune(p.code[p.cur:])
	return r
}

func (p *parser) next() rune {
	if p.cur >= len(p.code) {
		return eof
	}
	r, size := utf8.DecodeRune(p.code[p.cur:])
	p.cur += size
	p.col++
	if r == '\n' {
		p.col = 0
		p.line++
	}
	return r
}

func (p *parser) error(msg string) {
	p.err = fmt.Errorf("line %d col %d: %s", p.line+1, p.col+1, msg)
}

func (p *parser) errorf(format string, a ...interface{}) {
	p.error(fmt.Sprintf(format, a...))
}

type commandID int

const (
	cmdInvalid commandID = iota
	cmdTitle
	cmdMagnify2
	cmdMagnify3
	cmdMagnify4
	cmdImage
)

type command struct {
	id             commandID
	text           []byte
	trimWhiteSpace bool
}

func cmd(id commandID, text string, trimWS bool) command {
	return command{
		id:             id,
		text:           []byte(text),
		trimWhiteSpace: trimWS,
	}
}

var commands = []command{
	cmd(cmdTitle, "title", true),
	cmd(cmdMagnify2, "2x", false),
	cmd(cmdMagnify3, "3x", false),
	cmd(cmdMagnify4, "4x", false),
	cmd(cmdImage, "image", true),
}

func init() {
	// make sure the longest text item comes first, this way matching commands
	// during parsing will match the right one even if two commands share a
	// common prefix (e.g. "image" and "image_crop")
	sort.Sort(longestTextFirst(commands))
}

type longestTextFirst []command

func (c longestTextFirst) Len() int           { return len(c) }
func (c longestTextFirst) Less(i, j int) bool { return len(c[i].text) > len(c[j].text) }
func (c longestTextFirst) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
