package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// DSL — origin balance formula DSL (см. docs/research/origin-vs-nova/
// formula-dsl.md). Используется в na_construction.{prod_*, cons_*,
// charge_*}: varbinary(255) PHP-выражения, парсимые eval().
//
// Поддерживается subset PHP-выражений:
//
//   - литералы int/float
//   - переменные {level}, {basic}, {temp}, {tech=NN}, {building=NN}
//   - операторы: + - * / **
//   - унарный минус
//   - группировка (...)
//   - функции pow, floor, ceil, round, abs, min, max
//
// Неподдерживаемое (выбрасывает ParseError):
//   - PHP-функции вне белого списка (sqrt, log, ...)
//   - условные операторы (?:)
//   - присваивания, точки с запятой, statements
//
// Эталоны округления PHP eval -> Go (см. план 64 §«Известные риски»):
//   - floor(x)  = math.Floor(x)
//   - ceil(x)   = math.Ceil(x)
//   - round(x)  = PHP round half-away-from-zero (НЕ math.Round, который
//                 half-to-even в Go 1.10+; реализуем roundHalfAwayFromZero).
//   - abs(x)    = math.Abs(x)

// VarBinding — значения переменных DSL для одного evaluation-контекста.
//
// Если переменная не задана, evaluator возвращает ErrUndefinedVar.
// Динамические формулы (с {temp}, {tech=N}, {building=N}) детектируются
// по наличию таких ссылок в AST — статический импорт их пропускает,
// помечает building как HasDynamicProduction.
type VarBinding struct {
	Level    *int
	Basic    *float64
	Temp     *int
	Tech     map[int]int // tech_id → level
	Building map[int]int // building_id → level
}

// ParseError — синтаксическая ошибка DSL с позицией.
type ParseError struct {
	Source string
	Pos    int
	Msg    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("dsl parse error at %d in %q: %s", e.Pos, e.Source, e.Msg)
}

// EvalError — ошибка выполнения (например, отсутствует переменная).
type EvalError struct {
	Msg string
}

func (e *EvalError) Error() string { return "dsl eval error: " + e.Msg }

// expr — корневой узел AST.
type expr interface {
	eval(VarBinding) (float64, error)
	// usesVar — true если вычисление зависит от runtime-переменной
	// (temp/tech/building). Используется для классификации формул как
	// статических (можно предвычислить) vs динамических (требует Go-
	// функцию).
	usesVar(name string) bool
}

type numLit struct{ v float64 }

func (n *numLit) eval(VarBinding) (float64, error) { return n.v, nil }
func (*numLit) usesVar(string) bool                { return false }

type varRef struct{ name string }

func (v *varRef) eval(b VarBinding) (float64, error) {
	switch v.name {
	case "level":
		if b.Level == nil {
			return 0, &EvalError{Msg: "{level} not bound"}
		}
		return float64(*b.Level), nil
	case "basic":
		if b.Basic == nil {
			return 0, &EvalError{Msg: "{basic} not bound"}
		}
		return *b.Basic, nil
	case "temp":
		if b.Temp == nil {
			return 0, &EvalError{Msg: "{temp} not bound"}
		}
		return float64(*b.Temp), nil
	default:
		return 0, &EvalError{Msg: "unknown variable: " + v.name}
	}
}

func (v *varRef) usesVar(name string) bool { return v.name == name }

type techRef struct{ id int }

func (t *techRef) eval(b VarBinding) (float64, error) {
	if b.Tech == nil {
		return 0, &EvalError{Msg: fmt.Sprintf("{tech=%d} not bound", t.id)}
	}
	v, ok := b.Tech[t.id]
	if !ok {
		return 0, nil // tech не открыт = уровень 0
	}
	return float64(v), nil
}

func (t *techRef) usesVar(name string) bool { return name == "tech" }

type buildingRef struct{ id int }

func (br *buildingRef) eval(b VarBinding) (float64, error) {
	if b.Building == nil {
		return 0, &EvalError{Msg: fmt.Sprintf("{building=%d} not bound", br.id)}
	}
	v, ok := b.Building[br.id]
	if !ok {
		return 0, nil
	}
	return float64(v), nil
}

func (br *buildingRef) usesVar(name string) bool { return name == "building" }

type binOp struct {
	op       byte // '+' '-' '*' '/' '^' (последний = pow через **)
	lhs, rhs expr
}

func (b *binOp) eval(bind VarBinding) (float64, error) {
	l, err := b.lhs.eval(bind)
	if err != nil {
		return 0, err
	}
	r, err := b.rhs.eval(bind)
	if err != nil {
		return 0, err
	}
	switch b.op {
	case '+':
		return l + r, nil
	case '-':
		return l - r, nil
	case '*':
		return l * r, nil
	case '/':
		if r == 0 {
			return 0, &EvalError{Msg: "division by zero"}
		}
		return l / r, nil
	case '^':
		return math.Pow(l, r), nil
	}
	return 0, &EvalError{Msg: fmt.Sprintf("unknown binop %c", b.op)}
}

func (b *binOp) usesVar(name string) bool {
	return b.lhs.usesVar(name) || b.rhs.usesVar(name)
}

type unaryNeg struct{ child expr }

func (u *unaryNeg) eval(b VarBinding) (float64, error) {
	v, err := u.child.eval(b)
	if err != nil {
		return 0, err
	}
	return -v, nil
}

func (u *unaryNeg) usesVar(n string) bool { return u.child.usesVar(n) }

type call struct {
	name string
	args []expr
}

func (c *call) eval(b VarBinding) (float64, error) {
	switch c.name {
	case "pow":
		if len(c.args) != 2 {
			return 0, &EvalError{Msg: "pow expects 2 args"}
		}
		base, err := c.args[0].eval(b)
		if err != nil {
			return 0, err
		}
		exp, err := c.args[1].eval(b)
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	case "floor":
		v, err := evalSingle(c, b)
		if err != nil {
			return 0, err
		}
		return math.Floor(v), nil
	case "ceil":
		v, err := evalSingle(c, b)
		if err != nil {
			return 0, err
		}
		return math.Ceil(v), nil
	case "round":
		v, err := evalSingle(c, b)
		if err != nil {
			return 0, err
		}
		return roundHalfAwayFromZero(v), nil
	case "abs":
		v, err := evalSingle(c, b)
		if err != nil {
			return 0, err
		}
		return math.Abs(v), nil
	case "min":
		if len(c.args) < 1 {
			return 0, &EvalError{Msg: "min needs >= 1 arg"}
		}
		first, err := c.args[0].eval(b)
		if err != nil {
			return 0, err
		}
		acc := first
		for _, a := range c.args[1:] {
			v, err := a.eval(b)
			if err != nil {
				return 0, err
			}
			if v < acc {
				acc = v
			}
		}
		return acc, nil
	case "max":
		if len(c.args) < 1 {
			return 0, &EvalError{Msg: "max needs >= 1 arg"}
		}
		first, err := c.args[0].eval(b)
		if err != nil {
			return 0, err
		}
		acc := first
		for _, a := range c.args[1:] {
			v, err := a.eval(b)
			if err != nil {
				return 0, err
			}
			if v > acc {
				acc = v
			}
		}
		return acc, nil
	default:
		return 0, &EvalError{Msg: "unknown function: " + c.name}
	}
}

func (c *call) usesVar(n string) bool {
	for _, a := range c.args {
		if a.usesVar(n) {
			return true
		}
	}
	return false
}

func evalSingle(c *call, b VarBinding) (float64, error) {
	if len(c.args) != 1 {
		return 0, &EvalError{Msg: c.name + " expects 1 arg"}
	}
	return c.args[0].eval(b)
}

// roundHalfAwayFromZero — PHP round() default behaviour:
// 0.5 → 1, -0.5 → -1 (half away from zero). Go math.Round тоже так
// делает (с Go 1.10+), но явная реализация — для ясности и для случая
// если потребуется PHP_ROUND_HALF_EVEN-режим.
func roundHalfAwayFromZero(v float64) float64 {
	if v < 0 {
		return math.Ceil(v - 0.5)
	}
	return math.Floor(v + 0.5)
}

// Parse — корневая функция парсера. Возвращает AST или ParseError.
//
// Пустая строка / только whitespace → возвращает nil expr (caller
// должен интерпретировать как «формула отсутствует, использовать basic
// или 0»).
func Parse(src string) (expr, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, nil
	}
	p := &parser{src: src}
	p.advance()
	e, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.tok.kind != tokEOF {
		return nil, &ParseError{Source: src, Pos: p.pos, Msg: "unexpected trailing token: " + p.tok.lit}
	}
	return e, nil
}

// MustParse — для тестов: panic при ошибке.
func MustParse(src string) expr {
	e, err := Parse(src)
	if err != nil {
		panic(err)
	}
	return e
}

// EvalNumber — convenience: парсит и сразу вычисляет.
func EvalNumber(src string, b VarBinding) (float64, error) {
	e, err := Parse(src)
	if err != nil {
		return 0, err
	}
	if e == nil {
		return 0, nil
	}
	return e.eval(b)
}

// IsDynamic — возвращает true если формула зависит от {temp}, {tech=N}
// или {building=N}. Такие формулы импортёр НЕ предвычисляет в таблицы —
// помечает их as has_dynamic_production: true и оставляет реализацию
// функцией в Go (internal/origin/economy/).
func IsDynamic(src string) (bool, error) {
	e, err := Parse(src)
	if err != nil {
		return false, err
	}
	if e == nil {
		return false, nil
	}
	return e.usesVar("temp") || e.usesVar("tech") || e.usesVar("building"), nil
}

// --- lexer/parser ---

type tokKind int

const (
	tokEOF tokKind = iota
	tokNum
	tokIdent
	tokVar // {level} {basic} {temp} {tech=NN} {building=NN}
	tokLParen
	tokRParen
	tokComma
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokPow // **
)

type token struct {
	kind tokKind
	lit  string
	num  float64
	// extra — для {tech=N}/{building=N}: id; для {var} — имя
	extra string
	id    int
}

type parser struct {
	src string
	pos int
	tok token
}

func (p *parser) advance() {
	for p.pos < len(p.src) && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
	if p.pos >= len(p.src) {
		p.tok = token{kind: tokEOF}
		return
	}
	c := p.src[p.pos]
	switch {
	case c == '(':
		p.tok = token{kind: tokLParen, lit: "("}
		p.pos++
	case c == ')':
		p.tok = token{kind: tokRParen, lit: ")"}
		p.pos++
	case c == ',':
		p.tok = token{kind: tokComma, lit: ","}
		p.pos++
	case c == '+':
		p.tok = token{kind: tokPlus, lit: "+"}
		p.pos++
	case c == '-':
		p.tok = token{kind: tokMinus, lit: "-"}
		p.pos++
	case c == '*':
		if p.pos+1 < len(p.src) && p.src[p.pos+1] == '*' {
			p.tok = token{kind: tokPow, lit: "**"}
			p.pos += 2
		} else {
			p.tok = token{kind: tokStar, lit: "*"}
			p.pos++
		}
	case c == '/':
		p.tok = token{kind: tokSlash, lit: "/"}
		p.pos++
	case c == '{':
		p.lexVar()
	case isDigit(c) || c == '.':
		p.lexNumber()
	case isIdentStart(c):
		p.lexIdent()
	default:
		p.tok = token{kind: tokEOF, lit: string(c)}
	}
}

func isDigit(c byte) bool       { return c >= '0' && c <= '9' }
func isIdentStart(c byte) bool  { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' }
func isIdentCont(c byte) bool   { return isIdentStart(c) || isDigit(c) }

func (p *parser) lexNumber() {
	start := p.pos
	for p.pos < len(p.src) && (isDigit(p.src[p.pos]) || p.src[p.pos] == '.') {
		p.pos++
	}
	lit := p.src[start:p.pos]
	v, err := strconv.ParseFloat(lit, 64)
	if err != nil {
		p.tok = token{kind: tokEOF, lit: lit}
		return
	}
	p.tok = token{kind: tokNum, lit: lit, num: v}
}

func (p *parser) lexIdent() {
	start := p.pos
	for p.pos < len(p.src) && isIdentCont(p.src[p.pos]) {
		p.pos++
	}
	p.tok = token{kind: tokIdent, lit: p.src[start:p.pos]}
}

// lexVar парсит {level}, {basic}, {temp}, {tech=NN}, {building=NN}.
// Регистр имени игнорируется (origin DSL встречается mixed-case).
func (p *parser) lexVar() {
	start := p.pos
	p.pos++ // skip '{'
	nameStart := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '}' && p.src[p.pos] != '=' {
		p.pos++
	}
	name := strings.ToLower(strings.TrimSpace(p.src[nameStart:p.pos]))
	id := 0
	if p.pos < len(p.src) && p.src[p.pos] == '=' {
		p.pos++
		idStart := p.pos
		for p.pos < len(p.src) && p.src[p.pos] != '}' {
			p.pos++
		}
		idLit := strings.TrimSpace(p.src[idStart:p.pos])
		v, err := strconv.Atoi(idLit)
		if err != nil {
			p.tok = token{kind: tokEOF, lit: p.src[start:p.pos]}
			return
		}
		id = v
	}
	if p.pos >= len(p.src) || p.src[p.pos] != '}' {
		p.tok = token{kind: tokEOF, lit: p.src[start:p.pos]}
		return
	}
	p.pos++ // skip '}'
	p.tok = token{kind: tokVar, lit: p.src[start:p.pos], extra: name, id: id}
}

// Грамматика (precedence от низкой к высокой):
//
//   expr      = addExpr
//   addExpr   = mulExpr (('+'|'-') mulExpr)*
//   mulExpr   = powExpr (('*'|'/') powExpr)*
//   powExpr   = unaryExpr ('**' powExpr)?       // правая ассоциативность
//   unaryExpr = '-' unaryExpr | atom
//   atom      = NUM | VAR | '(' expr ')' | IDENT '(' arglist ')'
//   arglist   = expr (',' expr)*

func (p *parser) parseExpr() (expr, error) { return p.parseAdd() }

func (p *parser) parseAdd() (expr, error) {
	lhs, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.tok.kind == tokPlus || p.tok.kind == tokMinus {
		op := byte('+')
		if p.tok.kind == tokMinus {
			op = '-'
		}
		p.advance()
		rhs, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		lhs = &binOp{op: op, lhs: lhs, rhs: rhs}
	}
	return lhs, nil
}

func (p *parser) parseMul() (expr, error) {
	lhs, err := p.parsePow()
	if err != nil {
		return nil, err
	}
	for p.tok.kind == tokStar || p.tok.kind == tokSlash {
		op := byte('*')
		if p.tok.kind == tokSlash {
			op = '/'
		}
		p.advance()
		rhs, err := p.parsePow()
		if err != nil {
			return nil, err
		}
		lhs = &binOp{op: op, lhs: lhs, rhs: rhs}
	}
	return lhs, nil
}

func (p *parser) parsePow() (expr, error) {
	lhs, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	if p.tok.kind == tokPow {
		p.advance()
		rhs, err := p.parsePow() // right-assoc
		if err != nil {
			return nil, err
		}
		return &binOp{op: '^', lhs: lhs, rhs: rhs}, nil
	}
	return lhs, nil
}

func (p *parser) parseUnary() (expr, error) {
	if p.tok.kind == tokMinus {
		p.advance()
		child, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unaryNeg{child: child}, nil
	}
	return p.parseAtom()
}

func (p *parser) parseAtom() (expr, error) {
	switch p.tok.kind {
	case tokNum:
		v := p.tok.num
		p.advance()
		return &numLit{v: v}, nil
	case tokVar:
		v := p.tok
		p.advance()
		switch v.extra {
		case "level", "basic", "temp":
			return &varRef{name: v.extra}, nil
		case "tech":
			return &techRef{id: v.id}, nil
		case "building":
			return &buildingRef{id: v.id}, nil
		default:
			return nil, &ParseError{Source: p.src, Pos: p.pos, Msg: "unknown variable " + v.lit}
		}
	case tokLParen:
		p.advance()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokRParen {
			return nil, &ParseError{Source: p.src, Pos: p.pos, Msg: "expected ')'"}
		}
		p.advance()
		return e, nil
	case tokIdent:
		name := p.tok.lit
		p.advance()
		if p.tok.kind != tokLParen {
			return nil, &ParseError{Source: p.src, Pos: p.pos, Msg: "expected '(' after function " + name}
		}
		p.advance()
		args := []expr{}
		if p.tok.kind != tokRParen {
			for {
				e, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, e)
				if p.tok.kind == tokComma {
					p.advance()
					continue
				}
				break
			}
		}
		if p.tok.kind != tokRParen {
			return nil, &ParseError{Source: p.src, Pos: p.pos, Msg: "expected ')' in call"}
		}
		p.advance()
		return &call{name: name, args: args}, nil
	default:
		return nil, &ParseError{Source: p.src, Pos: p.pos, Msg: "unexpected token: " + p.tok.lit}
	}
}
