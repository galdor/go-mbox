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
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"
)

const HeaderDateFormat = "Mon Jan 02 15:04:05 -0700 2006"

func ParseMessagePrefix(data []byte) (id string, date time.Time, err error) {
	const prefix = "From "

	if !bytes.HasPrefix(data, []byte(prefix)) {
		err = fmt.Errorf("missing prefix %q", prefix)
		return
	}
	data = data[len(prefix):]

	idx := bytes.IndexByte(data, ' ')
	if idx == -1 {
		err = errors.New("missing ' ' after message id")
		return
	}
	id = string(data[:idx])
	data = data[idx+1:]

	date, err = time.Parse(HeaderDateFormat, string(data))
	if err != nil {
		err = fmt.Errorf("invalid date %q", string(data))
		return
	}

	return
}

type Message struct {
	Id   string
	Date time.Time
	Data []byte
}

func NewMessage(id string, date time.Time, data []byte) *Message {
	msg := &Message{
		Id:   id,
		Date: date,
		Data: data,
	}

	return msg
}

var FromLineRegexp = regexp.MustCompile("^>+From ")

func UnescapeMessageData(data []byte) []byte {
	var buf bytes.Buffer
	start := true

	for len(data) > 0 {
		if !start {
			idx := bytes.Index(data, []byte{'\r', '\n'})
			if idx == -1 {
				buf.Write(data)
				break
			}

			buf.Write(data[:idx+2])
			data = data[idx+2:]
		}

		start = false

		match := FromLineRegexp.Find(data)
		if match == nil {
			continue
		}

		buf.Write(match[1:]) // skip the first '>'
		data = data[len(match):]
	}

	return buf.Bytes()
}
