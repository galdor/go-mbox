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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/galdor/go-cmdline"
	"github.com/galdor/go-mbox"
)

type msgListEntry struct {
	*mbox.Message
	Msg         *mail.Message
	RFC3339Date string
}

func main() {
	cmdline := cmdline.New()

	cmdline.AddCommand("help", "print help and exit")
	cmdline.AddCommand("extract",
		"extract messages from a mbox file to a maildir")
	cmdline.AddCommand("list", "list messages in a mbox file")

	cmdline.Parse(os.Args)

	var cmdFn func([]string)

	cmdName := cmdline.CommandName()
	cmdArgs := cmdline.CommandArgumentsValues()

	switch cmdName {
	case "help":
		cmdline.PrintUsage(os.Stdout)
		os.Exit(0)
	case "extract":
		cmdFn = cmdExtract
	case "list":
		cmdFn = cmdList
	}

	cmdArgv0 := os.Args[0] + " " + cmdName
	cmdFn(append([]string{cmdArgv0}, cmdArgs...))
}

func cmdExtract(args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("path", "the mbox file")
	cmdline.AddOption("o", "output", "directory",
		"the directory to extract messages to")
	cmdline.SetOptionDefault("o", ".")
	cmdline.Parse(args)

	format := mbox.Mboxrd // TODO option
	path := cmdline.ArgumentValue("path")
	dirPath := cmdline.OptionValue("output")

	mbox, err := mbox.Open(path, format)
	if err != nil {
		die("cannot open mbox: %v", err)
	}
	defer mbox.Close()

	newDirPath := filepath.Join(dirPath, "new")
	if err := os.MkdirAll(newDirPath, 0750); err != nil {
		die("cannot create %s: %v", dirPath, err)
	}
	tmpDirPath := filepath.Join(dirPath, "tmp")
	if err := os.MkdirAll(tmpDirPath, 0750); err != nil {
		die("cannot create %s: %v", dirPath, err)
	}

	for {
		mboxMsg, err := mbox.Read()
		if err != nil {
			die("cannot read message: %v", err)
		} else if mboxMsg == nil {
			break
		}

		fileName := fmt.Sprintf("%s,S=%d:2,", mboxMsg.Id, len(mboxMsg.Data))
		tmpPath := filepath.Join(newDirPath, fileName)
		newPath := filepath.Join(newDirPath, fileName)

		if err := ioutil.WriteFile(tmpPath, mboxMsg.Data, 0640); err != nil {
			die("cannot write %s: %v", tmpPath, err)
		}

		if err := os.Rename(tmpPath, newPath); err != nil {
			die("cannot rename %s to %s: %v", tmpPath, newPath, err)
		}
	}
}

func cmdList(args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("path", "the mbox file")
	cmdline.AddOption("t", "template", "template",
		"the template used for each message line")
	cmdline.SetOptionDefault("t",
		`{{.Id}} {{.RFC3339Date}} {{index .Msg.Header.Subject 0}}`)
	cmdline.Parse(args)

	tplStr := cmdline.OptionValue("template")
	tpl, err := template.New("message").Parse(tplStr)
	if err != nil {
		die("cannot parse template string: %v", err)
	}

	format := mbox.Mboxrd // TODO option
	path := cmdline.ArgumentValue("path")

	mbox, err := mbox.Open(path, format)
	if err != nil {
		die("cannot open mbox: %v", err)
	}
	defer mbox.Close()

	for {
		mboxMsg, err := mbox.Read()
		if err != nil {
			die("cannot read message: %v", err)
		} else if mboxMsg == nil {
			break
		}

		msg, err := mail.ReadMessage(bytes.NewReader(mboxMsg.Data))
		if err != nil {
			warn("cannot parse message %s: %v", mboxMsg.Id, err)
			continue
		}

		entry := msgListEntry{
			Message:     mboxMsg,
			Msg:         msg,
			RFC3339Date: mboxMsg.Date.Format(time.RFC3339),
		}

		if err := tpl.Execute(os.Stdout, entry); err != nil {
			die("cannot execute template: %v", err)
		}
		fmt.Println()
	}
}

func warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func die(format string, args ...interface{}) {
	warn(format, args...)
	os.Exit(1)
}
