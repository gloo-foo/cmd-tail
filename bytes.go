package command

import (
	"bytes"
	"context"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
)

// newline is the line separator used to reconstitute and re-split byte-mode output.
var newline = []byte{'\n'}

// tailBytes reconstitutes the byte stream, keeps its final n bytes, and re-emits
// them as a line-oriented stream. An empty tail produces no output.
func tailBytes(n TailBytes) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.GenerateFrom(ctx, in, emitLastBytes(in, n))
	})
}

// emitLastBytes returns a producer that drains in, keeps the trailing n bytes
// of the reconstituted input, and re-emits them as a line-oriented stream.
func emitLastBytes(in gloo.Stream[[]byte], n TailBytes) func(context.Context, func([]byte) bool, func(error)) {
	return func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
		items, err := rill.ToSlice(in.Chan())
		if err != nil {
			sendErr(err)
			return
		}
		sendAll(send, splitLines(lastBytes(reconstitute(items), n)))
	}
}

// sendAll emits each line downstream, stopping early if send signals no more.
func sendAll(send func([]byte) bool, lines [][]byte) {
	for _, line := range lines {
		if !send(line) {
			return
		}
	}
}

// splitLines restores the line-oriented stream from the trailing byte buffer,
// dropping the empty trailer left by a terminating newline so each value is one
// line with its newline stripped. An empty buffer yields no lines.
func splitLines(buf []byte) [][]byte {
	if len(buf) == 0 {
		return nil
	}
	lines := bytes.Split(buf, newline)
	if last := len(lines) - 1; len(lines[last]) == 0 {
		return lines[:last]
	}
	return lines
}

// reconstitute rejoins stream items into a single newline-terminated buffer,
// the inverse of the line splitting that produced the stream.
func reconstitute(items [][]byte) []byte {
	var buf []byte
	for _, item := range items {
		buf = append(buf, item...)
		buf = append(buf, '\n')
	}
	return buf
}

// lastBytes returns the final n bytes of buf, or all of buf when it is shorter.
func lastBytes(buf []byte, n TailBytes) []byte {
	return buf[max(0, len(buf)-int(n)):]
}
