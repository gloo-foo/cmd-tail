// Package alias provides unprefixed type aliases for tail command flags.
//
//	import "github.com/gloo-foo/cmd-tail/alias"
//	tail.Tail(alias.Lines(5))
package alias

import (
	gloo "github.com/gloo-foo/framework"

	command "github.com/gloo-foo/cmd-tail"
)

// Tail returns a Command that outputs a trailing slice of its input.
func Tail(opts ...any) gloo.Command[[]byte, []byte] { return command.Tail(opts...) }

// Lines sets the number of trailing lines to output (-n flag).
type Lines = command.TailLines

// FromLine outputs every line from line N onward (-n +N flag).
type FromLine = command.TailFromLine

// Bytes sets the number of trailing bytes to output (-c flag).
type Bytes = command.TailBytes
