// Copyright (c) 2017 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package mbox

import (
	"testing"
)

func TestUnescapeMessageData(t *testing.T) {
	tests := []struct {
		escaped   string
		unescaped string
	}{
		{"",
			""},
		{"Foo\r\n",
			"Foo\r\n"},
		{"From \r\n",
			"From \r\n"},
		{">From \r\n",
			"From \r\n"},
		{">From foo\r\n",
			"From foo\r\n"},
		{">>From foo\r\n",
			">From foo\r\n"},
		{">>>From foo\r\n",
			">>From foo\r\n"},
		{"Foo\r\n>From foo\r\nBar\r\n",
			"Foo\r\nFrom foo\r\nBar\r\n"},
		{">From foo\r\nFoo\r\n>From bar\r\n",
			"From foo\r\nFoo\r\nFrom bar\r\n"},
		{">From foo\r\n\r\nFoo\r\n\r\n>From bar\r\n",
			"From foo\r\n\r\nFoo\r\n\r\nFrom bar\r\n"},
	}

	for _, test := range tests {
		unescaped := string(UnescapeMessageData([]byte(test.escaped)))
		if unescaped != test.unescaped {
			t.Errorf("%q was unescaped to %q instead of %q",
				test.escaped, unescaped, test.unescaped)
		}
	}
}
