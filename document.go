package main

type document struct {
	title string
	parts []docPart
}

type docPart interface {
	isDocPart()
}

type (
	docText          string
	docTitle         string
	docCaption       string
	docSubCaption    string
	docSubSubCaption string

	stylizedDocText struct {
		text         string
		bold, italic bool
	}

	docImage struct {
		name string
	}

	docLink struct {
		id   int
		text string
	}

	docLinkTarget int
)

func (docText) isDocPart()          {}
func (stylizedDocText) isDocPart()  {}
func (docImage) isDocPart()         {}
func (docLink) isDocPart()          {}
func (docLinkTarget) isDocPart()    {}
func (docTitle) isDocPart()         {}
func (docCaption) isDocPart()       {}
func (docSubCaption) isDocPart()    {}
func (docSubSubCaption) isDocPart() {}
