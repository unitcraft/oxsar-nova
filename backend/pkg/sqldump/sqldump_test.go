package sqldump

import "testing"

func TestParseInserts_Basic(t *testing.T) {
	t.Parallel()
	src := `
-- comment
INSERT INTO ` + "`t`" + ` (` + "`a`, `b`" + `) VALUES
(1, 'hello'),
(2, NULL);
`
	d, err := ParseInserts(src, "t")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(d.Rows) != 2 {
		t.Fatalf("rows: %d", len(d.Rows))
	}
	if d.Rows[1][1].IsNull != true {
		t.Fatalf("NULL not parsed")
	}
}

func TestParseInserts_EscapedQuote(t *testing.T) {
	t.Parallel()
	src := `INSERT INTO ` + "`t`" + ` (s) VALUES ('he said ''hi''');`
	d, err := ParseInserts(src, "t")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if d.Rows[0][0].Raw != "he said 'hi'" {
		t.Fatalf("got %q", d.Rows[0][0].Raw)
	}
}

func TestParseInserts_FormulaWithBraces(t *testing.T) {
	t.Parallel()
	src := `INSERT INTO ` + "`t`" + ` (f) VALUES ('floor({basic} * pow(1.5, ({level} - 1)))');`
	d, err := ParseInserts(src, "t")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if d.Rows[0][0].Raw != "floor({basic} * pow(1.5, ({level} - 1)))" {
		t.Fatalf("got %q", d.Rows[0][0].Raw)
	}
}

func TestParseInserts_AbsentTable(t *testing.T) {
	t.Parallel()
	d, err := ParseInserts("-- nothing", "absent")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(d.Rows) != 0 {
		t.Fatalf("expected no rows")
	}
}

func TestValue_String(t *testing.T) {
	t.Parallel()
	if v := (Value{IsNull: true}).String(); v != "NULL" {
		t.Errorf("null.String() = %q, want NULL", v)
	}
	if v := (Value{Raw: "hello"}).String(); v != "hello" {
		t.Errorf("raw.String() = %q, want hello", v)
	}
}

func TestValue_AsInt(t *testing.T) {
	t.Parallel()
	v := Value{Raw: "42"}
	n, err := v.AsInt()
	if err != nil || n != 42 {
		t.Errorf("AsInt() = %d, %v", n, err)
	}
	null := Value{IsNull: true}
	if n, _ := null.AsInt(); n != 0 {
		t.Errorf("null.AsInt() = %d, want 0", n)
	}
}

func TestValue_AsFloat(t *testing.T) {
	t.Parallel()
	v := Value{Raw: "3.14"}
	f, err := v.AsFloat()
	if err != nil || f != 3.14 {
		t.Errorf("AsFloat() = %v, %v", f, err)
	}
	null := Value{IsNull: true}
	if f, _ := null.AsFloat(); f != 0 {
		t.Errorf("null.AsFloat() = %v, want 0", f)
	}
}

func TestIndexColumns(t *testing.T) {
	t.Parallel()
	idx := IndexColumns([]string{"a", "b", "c"})
	if idx["a"] != 0 || idx["b"] != 1 || idx["c"] != 2 {
		t.Errorf("IndexColumns result: %v", idx)
	}
}

func TestAssignInt(t *testing.T) {
	t.Parallel()
	col := IndexColumns([]string{"id", "val"})
	row := []Value{{Raw: "7"}, {Raw: "42"}}
	var v int64
	if err := AssignInt(col, row, "val", &v); err != nil || v != 42 {
		t.Errorf("AssignInt: %d, %v", v, err)
	}
	if err := AssignInt(col, row, "missing", &v); err == nil {
		t.Error("expected error for missing column")
	}
}

func TestGetStr(t *testing.T) {
	t.Parallel()
	col := IndexColumns([]string{"name"})
	row := []Value{{Raw: "Alice"}}
	s, err := GetStr(col, row, "name")
	if err != nil || s != "Alice" {
		t.Errorf("GetStr = %q, %v", s, err)
	}
	if _, err := GetStr(col, row, "missing"); err == nil {
		t.Error("expected error for missing column")
	}
}
