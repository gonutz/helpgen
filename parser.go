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
	p.vars = make(map[string]string)
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
	vars      map[string]string
}

func (p *parser) parse() {
	for p.err == nil {
		r := p.next()
		if r == eof {
			p.finishText()
			return
		} else if r == '\\' {
			if p.peek() == '\\' {
				p.next()
				p.text += "\\"
			} else {
				p.finishText()
				p.parseCommand()
			}
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
			allParams := string(p.code[start : p.cur-1])
			params := strings.Split(allParams, ",")
			if len(params) != commands[i].paramCount {
				p.errorf(
					"wrong number of arguments for '%s' command, expected %d but have %d",
					commands[i].text,
					commands[i].paramCount,
					len(params),
				)
				return
			}

			switch commands[i].id {
			case cmdTitle:
				p.doc.title = strings.TrimSpace(params[0])
				p.eatWhiteSpace()
			case cmdImage:
				p.emit(docImage{name: strings.TrimSpace(params[0])})
			case cmdMagnify2:
				p.emit(largerDocText{text: params[0], scale: 2})
			case cmdMagnify3:
				p.emit(largerDocText{text: params[0], scale: 3})
			case cmdMagnify4:
				p.emit(largerDocText{text: params[0], scale: 4})
			case cmdSet:
				p.vars[strings.TrimSpace(params[0])] = params[1]
				p.eatWhiteSpace()
			case cmdVar:
				name := strings.TrimSpace(params[0])
				if value, ok := p.vars[name]; ok {
					p.text += value
				} else {
					p.errorf("undefined variable '%s': ", name)
					return
				}
			default:
				p.errorf("unhandled command: '%s'", commands[i].text)
			}
			return
		}
	}
	// If we came here, no known command was found. For a helpful error message,
	// name the command in the error message. NOTE that basically anything can
	// come after the \ so try to find the first non-character to mark the
	// supposed command's end
	cmd := ""
	rest := p.code[p.cur:]
	i := 0
	for i < len(rest) {
		r, size := utf8.DecodeRune(rest)
		rest = rest[size:]
		if !unicode.IsLetter(r) {
			break
		}
		if len(cmd) >= 30 {
			// do not show the whole rest of the code if no non-character is
			// found, clamp here
			cmd += "..."
			break
		}
		cmd += string(r)
	}
	p.errorf("unknown command: '%s'", cmd)
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
	p.err = fmt.Errorf("parse error in line %d col %d: %s", p.line+1, p.col+1, msg)
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
	cmdSet
	cmdVar
)

type command struct {
	id         commandID
	text       []byte
	paramCount int
}

func cmd(id commandID, text string, params int) command {
	return command{
		id:         id,
		text:       []byte(text),
		paramCount: params,
	}
}

var commands = []command{
	cmd(cmdTitle, "title", 1),
	cmd(cmdMagnify2, "2x", 1),
	cmd(cmdMagnify3, "3x", 1),
	cmd(cmdMagnify4, "4x", 1),
	cmd(cmdImage, "image", 1),
	cmd(cmdSet, "set", 2),
	cmd(cmdVar, "var", 1),
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
