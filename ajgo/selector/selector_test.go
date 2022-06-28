package selector

import (
	"testing"
)

var inputs = []string{
	`keep:seq:^chr`,
	`keep::^chr`,
	`keep:seq:`,
	`:seq:`,
	`::`,
	`::_:_`,
}

var expected = [][]string{
	{`keep`, `seq`, `^chr`},
	{`keep`, ``, `^chr`},
	{`keep`, `seq`, ``},
	{``, `seq`, ``},
	{``, ``, ``},
	{``, ``, `_:_`},
}

func TestNewFromString(t *testing.T) {
	for i, _ := range inputs {
		e := expected[i]
		g, err := NewFromString(inputs[i])
		if err != nil {
			t.Fatalf("unexpected error in NewSelector(): %v", err)
		}
		if e[0] != g.Operation {
			t.Fatalf(`Selector.Operation incorrect, should be %s but is %s`, e[0], g.Operation)
		}
		if e[1] != g.Subject {
			t.Fatalf(`Selector.Subject incorrect, should be %s but is %s`, e[1], g.Subject)
		}
		if e[2] != g.Pattern {
			t.Fatalf(`Selector.Pattern incorrect, should be %s but is %s`, e[2], g.Pattern)
		}
	}
}
