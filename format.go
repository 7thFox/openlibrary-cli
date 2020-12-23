package main

import (
	"reflect"
	"strings"
)

type CompiledFormat struct {
	InputFormat string

	write formatWriter
}

func (c *CompiledFormat) Format(i *BookInfo) string {
	var sb strings.Builder
	c.write(i, &sb)
	return sb.String()
}

func CompileFormat(format string) *CompiledFormat {
	c := CompiledFormat{
		InputFormat: format,
		write:       func(i *BookInfo, sb *strings.Builder) {},
	}

	state := parseStart(&c)
	reader := strings.NewReader(format)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		state = state(r)
	}
	state(eofSentinal)
	return &c
}

type formatWriter func(i *BookInfo, sb *strings.Builder)

func (c *CompiledFormat) appendToWrite(appended formatWriter) {
	before := c.write
	c.write = func(i *BookInfo, sb *strings.Builder) {
		before(i, sb)
		appended(i, sb)
	}
}

const (
	// startSentinal = rune(-999)
	eofSentinal = rune(-998)

// bindSentinal = rune(-997)
)

////////////////// PARSING ////////////////////////

type parseState func(r rune) parseState

func parseStart(c *CompiledFormat) parseState {
	return parseBindingOrLiteral(c)
}

func parseBindingOrLiteral(c *CompiledFormat) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			return nil
		}
		var literal strings.Builder
		if r == '{' {
			return parseBindingOrEscape(c, &literal)
		}

		literal.WriteRune(r)
		return parseLiteral(c, &literal)
	}
}

func parseLiteral(c *CompiledFormat, nextLiteral *strings.Builder) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			literal := nextLiteral.String()
			c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
				sb.WriteString(literal)
			})
			return nil
		}
		if r == '{' {
			return parseBindingOrEscape(c, nextLiteral)
		}
		nextLiteral.WriteRune(r)
		return parseLiteral(c, nextLiteral)
	}
}

func parseBindingOrEscape(c *CompiledFormat, leadingLiteral *strings.Builder) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			panic("Invalid Format: Unterminated bind")
		}
		if r == '{' { // Escape
			leadingLiteral.WriteRune('{')
			return parseLiteral(c, leadingLiteral)
		}
		if r == '}' || r == eofSentinal {
			colorPrintln(colorYellow, "WARN: Format contains unnecessary empty binding")
			return parseLiteral(c, leadingLiteral)
		}

		literal := leadingLiteral.String()
		c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
			sb.WriteString(literal)
		})

		var nextBind strings.Builder
		nextBind.WriteRune(r)
		return parseBinding(c, &nextBind)
	}
}

func parseBinding(c *CompiledFormat, nextBind *strings.Builder) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			panic("Invalid Format: Unterminated bind")
		}
		if r == '}' { // End (non-empty)
			bindParts := strings.Split(nextBind.String(), ".")
			return parseSmartSpaceBindingOrLiteral(
				c,
				func(i *BookInfo) string {
					return strings.TrimSpace(getField(reflect.ValueOf(*i), "", bindParts))
				})
		}
		nextBind.WriteRune(r)
		return parseBinding(c, nextBind)
	}
}

func parseSmartSpaceBindingOrLiteral(c *CompiledFormat, getBind func(i *BookInfo) string) parseState {
	if !GetSettings().SmartSpace {
		c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
			sb.WriteString(getBind(i))
		})
		return parseBindingOrLiteral(c)
	}

	return func(r rune) parseState {
		if r == '{' {
			return parseSmartSpaceBindingOrEscape(c, getBind)
		}

		c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
			sb.WriteString(getBind(i))
		})
		var nextLiteral strings.Builder
		nextLiteral.WriteRune(r)
		return parseLiteral(c, &nextLiteral)
	}
}

func parseSmartSpaceBindingOrEscape(c *CompiledFormat, getBind func(i *BookInfo) string) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			panic("Invalid Format: Unterminated bind")
		}
		if r == '{' {
			c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
				sb.WriteString(getBind(i))
			})
			var nextLiteral strings.Builder
			nextLiteral.WriteRune('{')
			return parseLiteral(c, &nextLiteral)
		}

		var nextBind strings.Builder
		nextBind.WriteRune(r)
		return parseSmartSpaceBinding(c, &nextBind, getBind)
	}
}

func parseSmartSpaceBinding(c *CompiledFormat, nextBind *strings.Builder, getBind func(i *BookInfo) string) parseState {
	return func(r rune) parseState {
		if r == eofSentinal {
			panic("Invalid Format: Unterminated bind")
		}
		if r == '}' {
			bindParts := strings.Split(nextBind.String(), ".")
			getRightBind := func(i *BookInfo) string {
				return strings.TrimSpace(getField(reflect.ValueOf(*i), "", bindParts))
			}
			c.appendToWrite(func(i *BookInfo, sb *strings.Builder) {
				left := getBind(i)
				if left != "" {
					sb.WriteString(left)
					if getRightBind(i) != "" { // This means we'll do our reflection call x2, which is mildly unfortunate
						sb.WriteRune(' ')
					}
				}
			})
			return parseSmartSpaceBindingOrLiteral(c, getRightBind)
		}

		nextBind.WriteRune(r)
		return parseSmartSpaceBinding(c, nextBind, getBind)
	}
}
