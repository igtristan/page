package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

type Tag struct {
	Type       string
	Attributes map[string]string
	Text       string
	Children   []*Tag
	// StartIndex - byte offset into the source file
	StartIndex int
	// Class - css class unique to this node
	Class string
	// ClassReferenced - was css generated for this class.  If so add the class to the tag
	ClassReferenced bool
}

func (t *Tag) GetChildrenWithType(ty string) []*Tag {
	out := []*Tag{}
	for _, a := range t.Children {
		if a.Type == ty {
			out = append(out, a)
		}
	}
	return out
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}

func isLowerAlpha(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func isLowerAlphaNum(c byte) bool {
	return ('a' <= c && c <= 'z') || ('0' <= c && c <= '9')
}

func isLowerAlphaNumDash(c byte) bool {
	return ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') || c == '-'
}

func isLowerAlphaNumDashDot(c byte) bool {
	return ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') || c == '-' || c == '.'
}
func parseFile(path string) (*Tag, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	br := &Reader{
		data: data,
	}

	tag := &Tag{Type: "*"}
	for {
		if err := parseWhitespace(br); err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			} else {
				break
			}
		}
		child, err := parseTag(br, false)
		if err != nil {
			return nil, err
		}
		tag.Children = append(tag.Children, child)
	}

	return tag, nil
}

func parseWhitespace(br *Reader) error {
	for {
		b, err := br.PeekByte()
		if err != nil || !isWhitespace(b) {
			return err
		}
		br.ReadByte()
	}
	return nil
}

func isAlpha(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isAlNum(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isAlNumDash(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '-'
}

func parseUnquotedAttributeName(br *Reader) (string, error) {

	buffer := bytes.Buffer{}
	for {
		peek, err := br.PeekByte()
		if err != nil {
			return "", err
		}

		if buffer.Len() == 0 && !isAlpha(peek) {
			return "", fmt.Errorf("attribute name must beging with a letter a-z")
		} else if isAlNumDash(peek) {
			br.ReadByte()
			buffer.WriteByte(peek)
		} else {
			break
		}
	}

	if buffer.Len() == 0 {
		return "", fmt.Errorf("No id found")
	}
	return buffer.String(), nil
}

func parseQuotedString(br *Reader) (string, error) {
	if err := br.Expect('"'); err != nil {
		return "", err
	}
	str, err := parseText(br, '"')
	if err != nil {
		return "", err
	}
	if err := br.Expect('"'); err != nil {
		return "", err
	}
	return str, nil
}

func parseText(br *Reader, term byte) (string, error) {
	buffer := bytes.Buffer{}
	for {
		tok, err := br.PeekByte()
		if err != nil {
			return "", err
		} else if tok == term {
			break
		}
		br.ReadByte()
		buffer.WriteByte(tok)
	}

	return buffer.String(), nil
}

func parseTag(br *Reader, startAlreadyConsumed bool) (*Tag, error) {

	//log.Println("START TAG")
	if !startAlreadyConsumed {
		if err := br.Expect('<'); err != nil {
			return nil, err
		}
	}

	startIndex := br.index - 1

	name := bytes.Buffer{}
	attributes := make(map[string]string)

	const modeTagName = 0
	const modeAttributes = 1
	const modeBody = 10
	const modeNonTagClosure = 11
	const modeComment = 12

	tag := &Tag{}
	text := bytes.Buffer{}

	mode := 0
	for {
		tok, err := br.PeekByte()
		if err != nil {
			return nil, err
		}

		// TAG NAME
		if mode == modeTagName {
			br.ReadByte()

			if tok == '!' {
				if err := br.ExpectString("--"); err != nil {
					return nil, err
				}
				name.WriteString("comment")
				mode = modeComment
			} else if isLowerAlphaNumDashDot(tok) {
				if name.Len() == 0 {
					if !isLowerAlpha(tok) {
						return nil, errors.New("tag name must start with [a-z]")
					}
				}
				name.WriteByte(tok)
			} else if isWhitespace(tok) && name.Len() > 0 {
				//log.Println("Read tag. type=", name.String())
				mode = modeAttributes
			} else if tok == '/' {
				mode = modeNonTagClosure
				break
			} else if tok == '>' && name.Len() > 0 {
				//log.Println("Read tag. type=", name.String())
				mode = modeBody
			} else {
				return nil, fmt.Errorf("Expecting tag name in lowercase: tn %v", string(tok))
			}
		} else if mode == modeComment {

			if tok == '>' && bytes.HasSuffix(text.Bytes(), []byte("--")) {
				mode = modeNonTagClosure
				break
			}

			br.ReadByte()
			text.WriteByte(tok)
		} else if mode == modeAttributes {
			// ATTRIBUTES
			if tok == '"' || isLowerAlpha(tok) {

				key := ""

				if tok == '"' {
					if key, err = parseQuotedString(br); err != nil {
						return nil, err
					}
				} else {
					if key, err = parseUnquotedAttributeName(br); err != nil {
						return nil, err
					}
				}
				if err := parseWhitespace(br); err != nil {
					return nil, err
				}
				if err := br.Expect('='); err != nil {
					return nil, err
				}
				if err := parseWhitespace(br); err != nil {
					return nil, err
				}
				value, err := parseQuotedString(br)
				if err != nil {
					return nil, err
				}

				attributes[key] = value
			} else if isWhitespace(tok) {
				br.ReadByte()
			} else if tok == '/' {
				br.ReadByte()
				mode = modeNonTagClosure
				break
			} else if tok == '>' {
				br.ReadByte()
				mode = modeBody
			} else {
				sline, scol := br.LineCol(startIndex)
				return nil, fmt.Errorf("%d:%d Invalid attribute in tag = %v", sline, scol, name.String())
			}
		} else if mode == modeBody {
			// BODY
			if isWhitespace(tok) {
				br.ReadByte()
				continue
			}

			if tok != '<' {
				// TEXT
				txt, err := parseText(br, '<')
				if err != nil {
					return nil, err
				}
				tag.Children = append(tag.Children, &Tag{
					Type: "text",
					Text: txt,
				})
			} else {
				// CHILD or CLOSE

				// break out if we've detected a closing tag
				br.ReadByte()
				if tok2, _ := br.PeekByte(); tok2 == '/' {
					break
				}

				// parse child tag
				child, err := parseTag(br, true)
				if err != nil {
					return nil, err
				}
				tag.Children = append(tag.Children, child)
			}
		}
	}

	if mode == modeNonTagClosure {
		if err := br.Expect('>'); err != nil {
			return nil, fmt.Errorf("Expecting early tag close />")
		}
	} else if err := br.ExpectString("/" + name.String() + ">"); err != nil {
		sline, scol := br.LineCol(startIndex)
		line, col := br.LineCol(br.index - 1)
		return nil, fmt.Errorf("%v:%v, expecting closing tag </%v> started at %v:%v", line, col, name.String(), sline, scol)
	}

	tag.StartIndex = startIndex
	tag.Attributes = attributes
	tag.Type = name.String()
	tag.Text = text.String()

	if tag.Type == "core.css" {
		if len(tag.Children) > 0 {
			tag.Text = tag.Children[0].Text
			tag.Children = tag.Children[:0]
		}
	}

	//log.Println("END TAG", tag.Type)
	return tag, nil
}
