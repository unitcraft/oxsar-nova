// Package sqldump — минимальный парсер phpMyAdmin SQL-дампов.
//
// Предназначен для чтения d:\Sources\oxsar2\sql\table_dump\*.sql и
// конвертации их в YAML (см. cmd/tools/import-datasheets и
// cmd/tools/import-phrases, §1.4 + §10.3 ТЗ).
//
// Поддерживаем:
//   - INSERT INTO `table` (col, col, ...) VALUES (v, v, ...), (...);
//   - одинарные кавычки с escape \' и \\ и двойной одинарной '';
//   - NULL без кавычек;
//   - числа (целые/float, с возможным -);
//   - строчные комментарии -- .
//
// Сознательно не поддерживаем:
//   - многострочные комментарии /* */;
//   - BLOB/HEX;
//   - INSERT без списка колонок;
//   - CREATE TABLE (только данные).
package sqldump

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Data — результат парсинга одной таблицы.
type Data struct {
	Table   string
	Columns []string
	Rows    [][]Value
}

// Value — одно SQL-значение.
//
// Инварианты:
//   - IsNull=true: остальные поля не значат.
//   - IsString=true: Raw — содержимое строки без кавычек и с
//     раскрытыми escape-последовательностями.
//   - иначе: Raw — лексема числа как строка.
type Value struct {
	IsNull   bool
	IsString bool
	Raw      string
}

func (v Value) String() string {
	if v.IsNull {
		return "NULL"
	}
	return v.Raw
}

// AsInt конвертирует значение в int64. NULL → 0.
func (v Value) AsInt() (int64, error) {
	if v.IsNull {
		return 0, nil
	}
	return strconv.ParseInt(strings.TrimSpace(v.Raw), 10, 64)
}

// AsFloat конвертирует значение в float64. NULL → 0.
func (v Value) AsFloat() (float64, error) {
	if v.IsNull {
		return 0, nil
	}
	return strconv.ParseFloat(strings.TrimSpace(v.Raw), 64)
}

// ParseInserts находит все INSERT INTO tableName и возвращает строки.
// Если таких INSERT нет — возвращает пустую Data без ошибки.
func ParseInserts(sql, tableName string) (Data, error) {
	out := Data{Table: tableName}
	idx := 0
	for {
		prefix := "INSERT INTO `" + tableName + "`"
		p := strings.Index(sql[idx:], prefix)
		if p < 0 {
			break
		}
		start := idx + p + len(prefix)

		openCol := strings.Index(sql[start:], "(")
		if openCol < 0 {
			return out, errors.New("sqldump: missing columns '(' after INSERT")
		}
		closeCol := strings.Index(sql[start+openCol:], ")")
		if closeCol < 0 {
			return out, errors.New("sqldump: missing columns ')' after INSERT")
		}
		columnsRaw := sql[start+openCol+1 : start+openCol+closeCol]
		columns := parseColumnList(columnsRaw)
		if len(out.Columns) == 0 {
			out.Columns = columns
		}

		after := start + openCol + closeCol + 1
		valuesKw := strings.Index(sql[after:], "VALUES")
		if valuesKw < 0 {
			return out, errors.New("sqldump: missing VALUES")
		}
		after += valuesKw + len("VALUES")

		rows, end, err := parseRowList(sql[after:])
		if err != nil {
			return out, fmt.Errorf("sqldump: parse rows: %w", err)
		}
		out.Rows = append(out.Rows, rows...)
		idx = after + end
	}
	return out, nil
}

// IndexColumns возвращает map имя→индекс для удобного доступа.
func IndexColumns(cols []string) map[string]int {
	m := make(map[string]int, len(cols))
	for i, c := range cols {
		m[c] = i
	}
	return m
}

// AssignInt читает колонку `name` как int64 в *ptr.
func AssignInt(col map[string]int, row []Value, name string, ptr *int64) error {
	i, ok := col[name]
	if !ok {
		return fmt.Errorf("column %q missing", name)
	}
	if i >= len(row) {
		return fmt.Errorf("row too short for %q", name)
	}
	v, err := row[i].AsInt()
	if err != nil {
		return fmt.Errorf("column %q: %w", name, err)
	}
	*ptr = v
	return nil
}

// GetStr возвращает .Raw значения колонки (пустая строка для NULL).
func GetStr(col map[string]int, row []Value, name string) (string, error) {
	i, ok := col[name]
	if !ok {
		return "", fmt.Errorf("column %q missing", name)
	}
	if i >= len(row) {
		return "", fmt.Errorf("row too short for %q", name)
	}
	return row[i].Raw, nil
}

// --- internal ---

func parseColumnList(src string) []string {
	parts := strings.Split(src, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "`")
		out = append(out, p)
	}
	return out
}

func parseRowList(src string) ([][]Value, int, error) {
	var rows [][]Value
	i := skipSpaces(src, 0)
	for i < len(src) {
		if src[i] == ';' {
			return rows, i + 1, nil
		}
		if src[i] != '(' {
			return nil, 0, fmt.Errorf("expected '(' at pos %d, got %q", i, src[i])
		}
		row, next, err := parseRow(src, i+1)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
		i = skipSpaces(src, next)
		if i < len(src) && src[i] == ',' {
			i = skipSpaces(src, i+1)
			continue
		}
		if i < len(src) && src[i] == ';' {
			return rows, i + 1, nil
		}
	}
	return nil, 0, errors.New("unterminated row list")
}

func parseRow(src string, start int) ([]Value, int, error) {
	var vals []Value
	i := skipSpaces(src, start)
	for i < len(src) {
		v, next, err := parseValue(src, i)
		if err != nil {
			return nil, 0, err
		}
		vals = append(vals, v)
		i = skipSpaces(src, next)
		if i >= len(src) {
			return nil, 0, errors.New("unterminated row")
		}
		if src[i] == ',' {
			i = skipSpaces(src, i+1)
			continue
		}
		if src[i] == ')' {
			return vals, i + 1, nil
		}
		return nil, 0, fmt.Errorf("unexpected char %q at row pos %d", src[i], i)
	}
	return nil, 0, errors.New("unterminated row")
}

func parseValue(src string, start int) (Value, int, error) {
	if start >= len(src) {
		return Value{}, 0, errors.New("unexpected EOF")
	}
	if src[start] == '\'' {
		return parseStringLiteral(src, start)
	}
	end := start
	for end < len(src) {
		c := src[end]
		if c == ',' || c == ')' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		end++
	}
	raw := src[start:end]
	upper := strings.ToUpper(strings.TrimSpace(raw))
	if upper == "NULL" {
		return Value{IsNull: true}, end, nil
	}
	return Value{Raw: raw}, end, nil
}

func parseStringLiteral(src string, start int) (Value, int, error) {
	var b strings.Builder
	i := start + 1
	for i < len(src) {
		c := src[i]
		if c == '\\' && i+1 < len(src) {
			next := src[i+1]
			switch next {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '\'':
				b.WriteByte('\'')
			case '"':
				b.WriteByte('"')
			case '0':
				b.WriteByte(0)
			default:
				b.WriteByte(next)
			}
			i += 2
			continue
		}
		if c == '\'' {
			if i+1 < len(src) && src[i+1] == '\'' {
				b.WriteByte('\'')
				i += 2
				continue
			}
			return Value{IsString: true, Raw: b.String()}, i + 1, nil
		}
		b.WriteByte(c)
		i++
	}
	return Value{}, 0, errors.New("unterminated string literal")
}

func skipSpaces(src string, i int) int {
	for i < len(src) {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if c == '-' && i+1 < len(src) && src[i+1] == '-' {
			for i < len(src) && src[i] != '\n' {
				i++
			}
			continue
		}
		break
	}
	return i
}
