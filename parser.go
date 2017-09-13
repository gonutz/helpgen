package main

import (
	"bytes"
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

func parse(code []byte) (document, error) {
	var p parser
	p.code = code
	p.parse()
	if p.err == nil {
		p.resolveRefs()
	}
	return p.doc, p.err
}

type parser struct {
	doc  document
	err  error
	code []byte
	vars varTable
}

type varTable map[string]variable

type variable struct {
	text           string
	declLineNumber int // 1-indexed
}

type codeLine struct {
	text   []byte
	kind   lineKind
	number int // 1-indexed
}

type lineKind int

const (
	textLine lineKind = iota
	equalsLine
	minusLine
	dottedLine
)

func (p *parser) parse() {
	p.code = unifyLineBreaks(p.code)
	lines := extractCodeLines(p.code)
	lines, p.vars, p.err = extractVariableDefinitions(lines)
	if p.err != nil {
		return
	}
	p.parseLines(lines)
	if p.err != nil {
		return
	}
	simplifyDoc(&p.doc)
}

func (p *parser) emit(part docPart) {
	p.doc.parts = append(p.doc.parts, part)
}

// unifyLineBreaks replaces all \r\n and \r with \n
func unifyLineBreaks(code []byte) []byte {
	code = bytes.Replace(code, []byte("\r\n"), []byte("\n"), -1)
	code = bytes.Replace(code, []byte("\r"), []byte("\n"), -1)
	return code
}

// extractCodeLines breaks the code up into lines and categorizes them
func extractCodeLines(code []byte) []codeLine {
	lineTexts := bytes.Split(code, []byte("\n"))
	lines := make([]codeLine, len(lineTexts))
	for i := range lines {
		lines[i].text = lineTexts[i]
		lines[i].kind = computeLineKind(lineTexts[i])
		lines[i].number = i + 1
	}
	return lines
}

func computeLineKind(line []byte) lineKind {
	if len(line) >= 3 {
		allSame := true
		for i := 1; i < len(line); i++ {
			if line[i] != line[i-1] {
				allSame = false
				break
			}
		}
		if allSame {
			switch line[0] {
			case '=':
				return equalsLine
			case '-':
				return minusLine
			case '.':
				return dottedLine
			}
		}
	}
	return textLine
}

func extractVariableDefinitions(lines []codeLine) ([]codeLine, varTable, error) {
	vars := make(varTable)
	varStart, varEnd := []byte(`[\`), []byte(`]`)
	eq := []byte("=")
	for i := 0; i < len(lines); i++ {
		line := lines[i].text
		if bytes.HasPrefix(line, varStart) && bytes.HasSuffix(line, varEnd) {
			firstEq := bytes.Index(line, eq)
			if firstEq >= 0 {
				name := string(line[len(varStart):firstEq])
				if validVarName(name) {
					// make sure each variable is only defined once
					if v, exists := vars[name]; exists {
						return nil, nil, fmt.Errorf(
							"variable '%s' redefined in line %d, first definition was in line %d, each variable can only be defined once",
							name,
							lines[i].number,
							v.declLineNumber,
						)
					}
					text := bytes.TrimSuffix(line[firstEq+1:], varEnd)
					v := variable{
						declLineNumber: lines[i].number,
						text:           string(text),
					}
					vars[name] = v
					// erase this line
					lines = append(lines[:i], lines[i+1:]...)
					i--
				}
			}
		}
	}
	return lines, vars, nil
}

// validVarName returns true if the name is not empty and contains only letters,
// digits or underscores
func validVarName(name string) bool {
	for _, r := range name {
		if !(r == ' ' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return name != ""
}

func simplifyDoc(doc *document) {
	// combine all neighbor pairs of docText into one
	for i := 0; i < len(doc.parts)-1; i++ {
		a, aIsText := doc.parts[i].(docText)
		b, bIsText := doc.parts[i+1].(docText)
		if aIsText && bIsText {
			doc.parts[i] = a + b
			doc.parts = append(doc.parts[:i+1], doc.parts[i+2:]...)
			i--
		}
	}
}

func (p *parser) parseLines(lines []codeLine) {
	// there can only be one title, having multiple titles is an error
	titleLine := -1
	for i, line := range lines {
		if line.kind == textLine {
			empty := len(bytes.TrimSpace(line.text)) == 0
			precededByEqualsLine := i > 0 && lines[i-1].kind == equalsLine
			followedByEqualsLine := i+1 < len(lines) && lines[i+1].kind == equalsLine
			followedByMinusLine := i+1 < len(lines) && lines[i+1].kind == minusLine
			followedByDottedLine := i+1 < len(lines) && lines[i+1].kind == dottedLine

			if !empty && precededByEqualsLine && followedByEqualsLine {
				// this is the document title, there can only be one
				if titleLine != -1 {
					p.err = fmt.Errorf("title redefined in line %d, first definition in line %d, there can only be one title", i+1, titleLine+1)
					return
				}
				titleLine = i
				p.doc.title = p.replaceVars(string(line.text))
				p.emit(docTitle(p.doc.title))
			} else if !empty && followedByEqualsLine {
				p.emit(docCaption(p.replaceVars(string(line.text))))
			} else if !empty && followedByMinusLine {
				p.emit(docSubCaption(p.replaceVars(string(line.text))))
			} else if !empty && followedByDottedLine {
				p.emit(docSubSubCaption(p.replaceVars(string(line.text))))
			} else {
				p.parseLine(line.text, line.number)
				if i != len(lines)-1 {
					p.emit(docText("\n"))
				}
			}
		}
	}
}

func (p *parser) parseLine(line []byte, lineNumber int) {
	if len(line) == 0 {
		return
	}

	i := 0
	for i < len(line) {
		switch line[i] {
		case '*', '/':
			delim := line[i]
			if i+1 < len(line) && !isSpace(line[i+1]) {
				end := findStyleEnd(line[i+1:], delim)
				if end != -1 {
					end += i + 1 // because index was for line[i+1:]
					if i > 0 {
						p.emit(docText(line[:i]))
					}
					text := string(line[i+1 : end-1])
					bold := delim == '*'
					italic := delim == '/'
					// account for both styles at once
					if bold && strings.HasPrefix(text, "/") && strings.HasSuffix(text, "/") {
						italic = true
						text = text[1 : len(text)-1]
					}
					if italic && strings.HasPrefix(text, "*") && strings.HasSuffix(text, "*") {
						bold = true
						text = text[1 : len(text)-1]
					}
					text = p.replaceVars(text)
					p.emit(stylizedDocText{
						bold:   bold,
						italic: italic,
						text:   text,
					})
					i = 0
					line = line[end:]
					continue
				}
			}
		case '[':
			ok, ref, subRef, rest := findRefEnd(line[i+1:])
			if ok {
				if i > 0 {
					p.emit(docText(line[:i]))
				}

				if subRef != "" {
					p.emit(tempRef{
						text:     ref,
						target:   subRef,
						declLine: lineNumber,
					})
				} else if v, ok := p.vars[ref]; ok {
					p.emit(docText(v.text))
				} else if len(ref) == 1 && strings.Contains("[*/=-.", ref) {
					p.emit(docText(ref))
				} else if hasImageExt(ref) {
					p.emit(docImage{name: ref})
				} else {
					p.emit(tempRef{
						target:   ref,
						declLine: lineNumber,
					})
				}

				i = 0
				line = rest
				continue
			}
		}
		i++
	}

	if len(line) > 0 {
		p.emit(docText(line))
	}
}

func (p *parser) replaceVars(text string) string {
	result := ""
	rest := text
	found := true
	for found {
		var next int
		rest, found, next = p.replaceFirstVar(rest)
		if !found {
			break
		}
		result += rest[:next]
		rest = rest[next:]
	}
	result += rest
	return result
}

func (p *parser) replaceFirstVar(text string) (string, bool, int) {
	for i, r := range text {
		if r == '[' {
			for j, s := range text[i+1:] {
				if s == ']' {
					name := text[i+1 : i+1+j]
					if v, ok := p.vars[name]; ok {
						return text[:i] + v.text + text[i+2+j:], true, i + len(v.text)
					}
				}
			}
		}
	}
	return text, false, 0
}

func (tempRef) isDocPart() {}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

func findStyleEnd(line []byte, delim byte) int {
	for i, b := range line {
		if b == delim && i > 0 && !isSpace(line[i-1]) {
			return i + 1
		}
	}
	return -1
}

func findRefEnd(line []byte) (found bool, ref, subRef string, rest []byte) {
	rest = line
	// no spaces at ref start and no empty references
	if len(line) == 0 || isSpace(line[0]) || line[0] == ']' {
		return
	}
	if len(line) >= 2 && line[0] == '[' && line[1] == ']' {
		// special case: "[[]" to escape s single '[' character
		found = true
		ref = "["
		rest = line[2:]
		return
	}
	// find the second ']', a sub-reference may be contained in this one
	subRefStart := -1
	for i, b := range line {
		if b == ']' {
			if subRefStart != -1 {
				// a sub reference must end with "]]"
				if i+1 < len(line) && line[i+1] == ']' {
					found = true
					ref = string(line[:subRefStart])
					subRef = string(line[subRefStart+1 : i])
					rest = line[i+2:]
					return
				} else {
					return
				}
			}
			found = true
			ref = string(line[:i])
			rest = line[i+1:]
			return
		}
		if b == '[' {
			if subRefStart != -1 {
				// only one extra '[' allowed in ref
				return
			}
			subRefStart = i
		}
	}
	return
}

func hasImageExt(s string) bool {
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".bmp", ".gif"} {
		if len(s) >= len(ext) && strings.ToLower(s[len(s)-len(ext):]) == ext {
			return true
		}
	}
	return false
}

type tempRef struct {
	declLine int
	target   string
	text     string
}

func (p *parser) resolveRefs() {
	// first find all referenced texts
	referenced := make(map[string]bool)
	for _, part := range p.doc.parts {
		if ref, ok := part.(tempRef); ok {
			referenced[ref.target] = true
		}
	}
	// add link targets for all referenced texts
	targets := make(map[string]int)
	i := 0
	addTarget := func(ref string) {
		if referenced[ref] {
			id := len(targets) + 1
			targets[ref] = id
			// insert this target into the document
			p.doc.parts = append(p.doc.parts, nil)
			copy(p.doc.parts[i+1:], p.doc.parts[i:])
			p.doc.parts[i] = docLinkTarget(id)
			i++
		}
	}
	for i < len(p.doc.parts) {
		part := p.doc.parts[i]
		if title, ok := part.(docTitle); ok {
			addTarget(string(title))
		}
		if caption, ok := part.(docCaption); ok {
			addTarget(string(caption))
		}
		if caption, ok := part.(docSubCaption); ok {
			addTarget(string(caption))
		}
		if caption, ok := part.(docSubSubCaption); ok {
			addTarget(string(caption))
		}
		i++
	}
	// replace all tempRefs with actual references
	for i, part := range p.doc.parts {
		if ref, ok := part.(tempRef); ok {
			target := targets[ref.target]
			if target == 0 {
				// in this case, check if we have a URL
				if strings.HasPrefix(ref.target, "www.") ||
					strings.HasPrefix(ref.target, "http://www.") ||
					strings.HasPrefix(ref.target, "https://www.") {
					text := ref.text
					if text == "" {
						text = ref.target
					}
					url := ref.target
					if strings.HasPrefix(url, "www.") {
						url = "http://" + url
					}
					p.doc.parts[i] = externalDocLink{
						url:  url,
						text: text,
					}
					continue
				}
				// see if this is a mail address
				possibleAddr := strings.TrimPrefix(ref.target, "mailto:")
				if addr, err := mail.ParseAddress(possibleAddr); err == nil {
					text := ref.text
					if text == "" {
						text = addr.Address
					}
					p.doc.parts[i] = externalDocLink{
						url:  "mailto:" + addr.Address,
						text: text,
					}
					continue
				}
				// neither a known internal link target nor a valid external
				// link -> error
				p.err = fmt.Errorf(
					"unknown link target '%s' in line %d",
					ref.target,
					ref.declLine,
				)
				return
			}
			text := ref.text
			if text == "" {
				text = ref.target
			}
			p.doc.parts[i] = docLink{
				id:   targets[ref.target],
				text: text,
			}
			referenced[ref.text] = true
		}
	}
}
