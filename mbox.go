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
	"errors"
	"fmt"
	"os"

	"github.com/galdor/go-stream"
)

type Format string

const (
	Mboxrd Format = "mboxrd"
)

func (f *Format) Parse(s string) error {
	switch s {
	case "mboxrd":
		*f = Mboxrd

	default:
		return errors.New("unknown format")
	}

	return nil
}

type Mbox struct {
	Format Format

	file   *os.File
	stream *stream.Stream
}

func Open(path string, format Format) (*Mbox, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	mbox := &Mbox{
		Format: format,

		file:   file,
		stream: stream.NewStream(file),
	}

	return mbox, nil
}

func (mbox *Mbox) Read() (*Message, error) {
	// Header
	line, err := mbox.stream.ReadUntilAndSkip([]byte{'\r', '\n'})
	if err != nil {
		return nil, err
	} else if line == nil {
		return nil, nil
	}

	id, date, err := ParseMessagePrefix(line)
	if err != nil {
		return nil, fmt.Errorf("invalid message header: %v", err)
	}

	// Message
	data, err := mbox.stream.ReadUntil([]byte("\r\nFrom "))
	if err != nil {
		return nil, fmt.Errorf("cannot read message %q: %v", id, err)
	}

	if data == nil {
		// Last message
		data, err = mbox.stream.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("cannot read last message %q: %v", id, err)
		}
	} else {
		err = mbox.stream.Skip(2) // "\r\n"
		if err != nil {
			return nil, fmt.Errorf("cannot skip end of message header %q: %v",
				id, err)
		}
	}

	msg := NewMessage(id, date, UnescapeMessageData(data))

	return msg, nil
}

func (mbox *Mbox) Close() {
	mbox.file.Close()
	mbox.file = nil
	mbox.stream = nil
}
