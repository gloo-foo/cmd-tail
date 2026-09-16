package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

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
	f := fold(opts)
	switch {
	case f.bytes > 0:
		return tailBytes(f.bytes)
	case f.fromLine > 0:
		return fromLine(f.fromLine)
	default:
		return lastLines(lineCount(f.lines))
	}
}

// lineCount resolves a TailLines flag to a concrete count, applying the
// ten-line default whenever the flag is unset or non-positive.
func lineCount(n TailLines) TailLines {
	if n <= 0 {
		return defaultLines
	}
	return n
}

// lastLines emits the final n lines of the accumulated input.
func lastLines(n TailLines) gloo.Command[[]byte, []byte] {
	return patterns.Accumulate(func(lines [][]byte) ([][]byte, error) {
		return lines[max(0, len(lines)-int(n)):], nil
	})
}

// fromLine emits every line from the 1-indexed position n onward (GNU -n +N).
func fromLine(n TailFromLine) gloo.Command[[]byte, []byte] {
	return patterns.Accumulate(func(lines [][]byte) ([][]byte, error) {
		return lines[min(int(n)-1, len(lines)):], nil
	})
}
