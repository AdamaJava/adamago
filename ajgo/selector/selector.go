package selector

import (
	"fmt"
	"strings"
)

// A Selector defines a selection operation, the type of thing to be
// selected, and the pattern (regex) to use for selection.
type Selector struct {
	Operation string
	Subject   string
	Pattern   string
}

func (s Selector) String() string {
	return s.Operation + ":" +
		s.Subject + ":" +
		s.Pattern
}

// NewFromString takes a string in the format operation:subject:pattern,
// parses it into a Selector and returns a pointer.
func NewFromString(s string) (*Selector, error) {
	var sel Selector
	ss := strings.SplitN(s, `:`, 3)
	if len(ss) != 3 {
		return &sel, fmt.Errorf("incorrectly formed selector: %s", s)
	}
	sel = Selector{
		Operation: ss[0],
		Subject:   ss[1],
		Pattern:   ss[2],
	}
	return &sel, nil
}

// NewFromStrings takes a list of strings in the format operation:subject:pattern
// and parses them into a list of pointers to Selector.
func NewFromStrings(selects []string) ([]*Selector, error) {
	var sels []*Selector
	for _, s := range selects {
		ss := strings.SplitN(s, `:`, 3)
		if len(ss) != 3 {
			return sels, fmt.Errorf("incorrectly formed selector: %s", s)
		}
		sels = append(sels,
			&Selector{Operation: ss[0],
				Subject: ss[1],
				Pattern: ss[2]})
	}
	return sels, nil
}
