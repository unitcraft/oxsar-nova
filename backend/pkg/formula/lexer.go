package formula

import (
	"fmt"
	"strings"
	"unicode"
)

// tokenKind — типы токенов.
type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNumber
	tokIdent
	tokVarLevel       // {level}
	tokVarBasic       // {basic}
	tokVarTemp        // {temp}
	tokVarTech        // {tech=N} — N хранится в value (как строка)
	tokLParen
	tokRParen
	tokComma
	tokPlus
	tokMinus
	tokStar
	tokSlash
)

// token — лексема с позицией (позиция пригодится в ошибках парсера).
type token struct {
	kind  tokenKind
	value string
	pos   int
}

// lex превращает входную строку в слайс токенов.
// Пробелы игнорируются. Неизвестный символ — ошибка.
func lex(src string) ([]token, error) {
	var out []token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '(':
			out = append(out, token{tokLParen, "(", i})
			i++
		case c == ')':
			out = append(out, token{tokRParen, ")", i})
			i++
		case c == ',':
			out = append(out, token{tokComma, ",", i})
			i++
		case c == '+':
			out = append(out, token{tokPlus, "+", i})
			i++
		case c == '-':
			out = append(out, token{tokMinus, "-", i})
			i++
		case c == '*':
			out = append(out, token{tokStar, "*", i})
			i++
		case c == '/':
			out = append(out, token{tokSlash, "/", i})
			i++
		case c == '{':
			tok, next, err := lexVar(src, i)
			if err != nil {
				return nil, err
			}
			out = append(out, tok)
			i = next
		case c == '.' || isDigit(c):
			tok, next := lexNumber(src, i)
			out = append(out, tok)
			i = next
		case isAlpha(c):
			tok, next := lexIdent(src, i)
			out = append(out, tok)
			i = next
		default:
			return nil, fmt.Errorf("formula: unexpected character %q at %d", c, i)
		}
	}
	out = append(out, token{tokEOF, "", len(src)})
	return out, nil
}

func lexVar(src string, start int) (token, int, error) {
	end := strings.IndexByte(src[start:], '}')
	if end < 0 {
		return token{}, 0, fmt.Errorf("formula: unterminated '{' at %d", start)
	}
	content := strings.TrimSpace(src[start+1 : start+end])
	next := start + end + 1
	switch {
	case content == "level":
		return token{tokVarLevel, "level", start}, next, nil
	case content == "basic":
		return token{tokVarBasic, "basic", start}, next, nil
	case content == "temp":
		return token{tokVarTemp, "temp", start}, next, nil
	case strings.HasPrefix(content, "tech="):
		id := strings.TrimSpace(strings.TrimPrefix(content, "tech="))
		if id == "" {
			return token{}, 0, fmt.Errorf("formula: empty tech id at %d", start)
		}
		// Проверка, что id — целое. Парсить будем в parser'е, чтобы
		// ошибка была с контекстом.
		for _, r := range id {
			if !unicode.IsDigit(r) {
				return token{}, 0, fmt.Errorf("formula: non-numeric tech id %q at %d", id, start)
			}
		}
		return token{tokVarTech, id, start}, next, nil
	default:
		return token{}, 0, fmt.Errorf("formula: unknown variable {%s} at %d", content, start)
	}
}

func lexNumber(src string, start int) (token, int) {
	i := start
	hasDot := false
	for i < len(src) {
		c := src[i]
		if isDigit(c) {
			i++
			continue
		}
		if c == '.' && !hasDot {
			hasDot = true
			i++
			continue
		}
		break
	}
	return token{tokNumber, src[start:i], start}, i
}

func lexIdent(src string, start int) (token, int) {
	i := start
	for i < len(src) && (isAlpha(src[i]) || isDigit(src[i]) || src[i] == '_') {
		i++
	}
	return token{tokIdent, src[start:i], start}, i
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' }
