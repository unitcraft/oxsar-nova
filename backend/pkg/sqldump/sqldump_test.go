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
