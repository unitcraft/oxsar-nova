package formula

import (
	"fmt"
	"strconv"
)

// Expr — корень AST. Exported для prefetch'а и кеша.
type Expr struct {
	node exprNode
}

// exprNode — внутренний узел AST.
type exprNode interface {
	eval(c Context) (float64, error)
}

// Parse разбирает строку формулы в AST. Пустая строка — nil-expr,
// Eval даст 0: так в legacy пустые prod_* означают «ничего не
// производит».
func Parse(src string) (*Expr, error) {
	if src == "" {
		return &Expr{node: numberNode{v: 0}}, nil
	}
	toks, err := lex(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	n, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.cur().kind != tokEOF {
		return nil, fmt.Errorf("formula: trailing tokens at %d", p.cur().pos)
	}
	return &Expr{node: n}, nil
}

// Eval вычисляет AST в контексте. Возвращает float64; вызывающий
// может взять math.Floor/Ceil, если хочет int.
func (e *Expr) Eval(c Context) (float64, error) {
	if e == nil || e.node == nil {
		return 0, nil
	}
	return e.node.eval(c)
}

// --- Парсер рекурсивного спуска ---

type parser struct {
	toks []token
	i    int
}

func (p *parser) cur() token       { return p.toks[p.i] }
func (p *parser) advance() token   { t := p.toks[p.i]; p.i++; return t }
func (p *parser) peek(k tokenKind) bool { return p.cur().kind == k }

func (p *parser) expect(k tokenKind, what string) (token, error) {
	if !p.peek(k) {
		return token{}, fmt.Errorf("formula: expected %s at %d, got %q", what, p.cur().pos, p.cur().value)
	}
	return p.advance(), nil
}

// parseExpr = parseTerm (('+'|'-') parseTerm)*
func (p *parser) parseExpr() (exprNode, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.peek(tokPlus) || p.peek(tokMinus) {
		op := p.advance().kind
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = binOpNode{op: op, l: left, r: right}
	}
	return left, nil
}

// parseTerm = parseUnary (('*'|'/') parseUnary)*
func (p *parser) parseTerm() (exprNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek(tokStar) || p.peek(tokSlash) {
		op := p.advance().kind
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = binOpNode{op: op, l: left, r: right}
	}
	return left, nil
}

// parseUnary = ('-' | '+') parseUnary | parsePrimary
func (p *parser) parseUnary() (exprNode, error) {
	if p.peek(tokMinus) {
		p.advance()
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return unaryNode{op: tokMinus, x: v}, nil
	}
	if p.peek(tokPlus) {
		p.advance()
		return p.parseUnary()
	}
	return p.parsePrimary()
}

// parsePrimary = Number | Var | Ident '(' args ')' | '(' Expr ')'
func (p *parser) parsePrimary() (exprNode, error) {
	t := p.cur()
	switch t.kind {
	case tokNumber:
		p.advance()
		v, err := strconv.ParseFloat(t.value, 64)
		if err != nil {
			return nil, fmt.Errorf("formula: bad number %q at %d: %w", t.value, t.pos, err)
		}
		return numberNode{v: v}, nil

	case tokVarLevel:
		p.advance()
		return varLevelNode{}, nil
	case tokVarBasic:
		p.advance()
		return varBasicNode{}, nil
	case tokVarTemp:
		p.advance()
		return varTempNode{}, nil
	case tokVarTech:
		p.advance()
		id, err := strconv.Atoi(t.value)
		if err != nil {
			return nil, fmt.Errorf("formula: bad tech id %q at %d", t.value, t.pos)
		}
		return varTechNode{id: id}, nil

	case tokLParen:
		p.advance()
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tokRParen, "')'"); err != nil {
			return nil, err
		}
		return inner, nil

	case tokIdent:
		p.advance()
		if _, err := p.expect(tokLParen, "'(' after function name"); err != nil {
			return nil, err
		}
		var args []exprNode
		if !p.peek(tokRParen) {
			for {
				a, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, a)
				if p.peek(tokComma) {
					p.advance()
					continue
				}
				break
			}
		}
		if _, err := p.expect(tokRParen, "')' after function args"); err != nil {
			return nil, err
		}
		fn, err := resolveFunc(t.value, len(args))
		if err != nil {
			return nil, fmt.Errorf("formula: %w at %d", err, t.pos)
		}
		return funcCallNode{fn: fn, name: t.value, args: args}, nil

	default:
		return nil, fmt.Errorf("formula: unexpected token %q at %d", t.value, t.pos)
	}
}
