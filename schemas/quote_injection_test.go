// Copyright 2024 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemas

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// quoteWordTo wraps an identifier in the dialect delimiter but wrote the body
// verbatim, so a delimiter INSIDE the identifier closed the quote the writer
// opened and the remainder executed. Desc("name`,(subquery)--") became
//
//	`name`,(subquery)--`
//
// two quoted identifiers plus bare SQL, not one identifier. The escape is to
// DOUBLE the closing delimiter (standard SQL identifier escaping); this asserts
// the emitted string, so it fails if the escape is dropped.
func TestQuoteWordCannotBreakOutOfItsDelimiter(t *testing.T) {
	cases := []struct {
		name   string
		quoter Quoter
		in     string
		want   string
	}{
		{
			name:   "backtick (mysql/sqlite)",
			quoter: Quoter{'`', '`', AlwaysReserve},
			in:     "name`,(select/**/token/**/from/**/secret)--",
			want:   "`name``,(select/**/token/**/from/**/secret)--`",
		},
		{
			name:   "double-quote (postgres/standard)",
			quoter: Quoter{'"', '"', AlwaysReserve},
			in:     `name",(select/**/token/**/from/**/secret)--`,
			want:   `"name"",(select/**/token/**/from/**/secret)--"`,
		},
		{
			name:   "bracket (mssql) — only the closing ] is doubled",
			quoter: Quoter{'[', ']', AlwaysReserve},
			in:     `name],(select/**/token/**/from/**/secret)--`,
			want:   `[name]],(select/**/token/**/from/**/secret)--]`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.quoter.Quote(c.in)
			t.Logf("Quote(%q) = %q", c.in, got)
			assert.Equal(t, c.want, got)

			// Structural invariant: strip the outer delimiter pair; every closing
			// delimiter left in the body must be doubled, so its count is even.
			// A lone one would terminate the quote and start bare SQL.
			inner := got[1 : len(got)-1]
			assert.Equal(t, 0, strings.Count(inner, string(c.quoter.Suffix))%2,
				"interior closing delimiters must all be escaped (doubled)")
		})
	}
}
