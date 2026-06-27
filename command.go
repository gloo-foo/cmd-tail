package command

import (
	"bytes"
	"context"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// newline is the line separator used to reconstitute and re-split byte-mode output.
var newline = []byte{'\n'}

// defaultLines is the number of trailing lines tail emits when no count flag is
// given, matching GNU tail's default of ten.
const defaultLines = 10

// Tail returns a Command that emits a trailing slice of its input.
//
// Modes (first match wins):
//   - TailBytes(n) (-c n): emit the last n bytes of the reconstituted input.
//   - TailFromLine(n) (-n +n): emit every line from the 1-indexed line n onward.
//   - TailLines(n) (-n n): emit the last n lines. n <= 0 falls back to ten.
//
// With no flag, Tail emits the last ten lines.
func Tail(opts ...any) gloo.Command[[]byte, []byte] {
	f := gloo.NewParameters[gloo.File, flags](opts...).Flags
	switch {
	case f.bytes > 0:
		return tailBytes(int(f.bytes))
	case f.fromLine > 0:
		return fromLine(int(f.fromLine))
	default:
		return lastLines(lineCount(f.lines))
	}
}

// lineCount resolves a TailLines flag to a concrete count, applying the
// ten-line default whenever the flag is unset or non-positive.
func lineCount(n TailLines) int {
	if n <= 0 {
		return defaultLines
	}
	return int(n)
}

// lastLines emits the final n lines of the accumulated input.
func lastLines(n int) gloo.Command[[]byte, []byte] {
	return patterns.Accumulate(func(lines [][]byte) ([][]byte, error) {
		return lines[max(0, len(lines)-n):], nil
	})
}

// fromLine emits every line from the 1-indexed position n onward (GNU -n +N).
func fromLine(n int) gloo.Command[[]byte, []byte] {
	return patterns.Accumulate(func(lines [][]byte) ([][]byte, error) {
		return lines[min(n-1, len(lines)):], nil
	})
}

// tailBytes reconstitutes the byte stream, keeps its final n bytes, and re-emits
// them as a line-oriented stream. An empty tail produces no output.
func tailBytes(n int) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.GenerateFrom(ctx, in, emitLastBytes(in, n))
	})
}

// emitLastBytes returns a producer that drains in, keeps the trailing n bytes
// of the reconstituted input, and re-emits them as a line-oriented stream.
func emitLastBytes(in gloo.Stream[[]byte], n int) func(context.Context, func([]byte) bool, func(error)) {
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
func lastBytes(buf []byte, n int) []byte {
	return buf[max(0, len(buf)-n):]
}
