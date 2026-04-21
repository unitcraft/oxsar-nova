package formula

import (
	"fmt"
	"math"
)

// numberNode — литерал.
type numberNode struct{ v float64 }

func (n numberNode) eval(Context) (float64, error) { return n.v, nil }

// varLevelNode / varBasicNode / varTempNode / varTechNode — переменные.
type varLevelNode struct{}

func (varLevelNode) eval(c Context) (float64, error) { return float64(c.Level), nil }

type varBasicNode struct{}

func (varBasicNode) eval(c Context) (float64, error) { return float64(c.Basic), nil }

type varTempNode struct{}

func (varTempNode) eval(c Context) (float64, error) { return float64(c.Temperature), nil }

type varTechNode struct{ id int }

func (n varTechNode) eval(c Context) (float64, error) { return float64(c.techLevel(n.id)), nil }

// binOpNode — бинарный оператор.
type binOpNode struct {
	op tokenKind
	l  exprNode
	r  exprNode
}

func (n binOpNode) eval(c Context) (float64, error) {
	lv, err := n.l.eval(c)
	if err != nil {
		return 0, err
	}
	rv, err := n.r.eval(c)
	if err != nil {
		return 0, err
	}
	switch n.op {
	case tokPlus:
		return lv + rv, nil
	case tokMinus:
		return lv - rv, nil
	case tokStar:
		return lv * rv, nil
	case tokSlash:
		if rv == 0 {
			return 0, fmt.Errorf("formula: division by zero")
		}
		return lv / rv, nil
	}
	return 0, fmt.Errorf("formula: unknown binary op")
}

// unaryNode — унарный минус.
type unaryNode struct {
	op tokenKind
	x  exprNode
}

func (n unaryNode) eval(c Context) (float64, error) {
	v, err := n.x.eval(c)
	if err != nil {
		return 0, err
	}
	if n.op == tokMinus {
		return -v, nil
	}
	return v, nil
}

// funcCallNode — вызов whitelist-функции.
type funcCallNode struct {
	fn   funcImpl
	name string
	args []exprNode
}

func (n funcCallNode) eval(c Context) (float64, error) {
	vs := make([]float64, len(n.args))
	for i, a := range n.args {
		v, err := a.eval(c)
		if err != nil {
			return 0, err
		}
		vs[i] = v
	}
	return n.fn(vs)
}

// funcImpl — реализация встроенной функции.
type funcImpl func(args []float64) (float64, error)

// resolveFunc находит функцию по имени с проверкой арности.
// Whitelist: floor, ceil, round, pow, min, max, sqrt, abs.
// Никакие другие имена не допускаются.
func resolveFunc(name string, arity int) (funcImpl, error) {
	switch name {
	case "floor":
		if arity != 1 {
			return nil, fmt.Errorf("floor: expected 1 arg, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Floor(a[0]), nil }, nil
	case "ceil":
		if arity != 1 {
			return nil, fmt.Errorf("ceil: expected 1 arg, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Ceil(a[0]), nil }, nil
	case "round":
		if arity != 1 {
			return nil, fmt.Errorf("round: expected 1 arg, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Round(a[0]), nil }, nil
	case "pow":
		if arity != 2 {
			return nil, fmt.Errorf("pow: expected 2 args, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Pow(a[0], a[1]), nil }, nil
	case "min":
		if arity != 2 {
			return nil, fmt.Errorf("min: expected 2 args, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Min(a[0], a[1]), nil }, nil
	case "max":
		if arity != 2 {
			return nil, fmt.Errorf("max: expected 2 args, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Max(a[0], a[1]), nil }, nil
	case "sqrt":
		if arity != 1 {
			return nil, fmt.Errorf("sqrt: expected 1 arg, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Sqrt(a[0]), nil }, nil
	case "abs":
		if arity != 1 {
			return nil, fmt.Errorf("abs: expected 1 arg, got %d", arity)
		}
		return func(a []float64) (float64, error) { return math.Abs(a[0]), nil }, nil
	}
	return nil, fmt.Errorf("unknown function %q", name)
}
