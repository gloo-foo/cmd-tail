package command_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
	"github.com/gloo-foo/testable/assertion"
	"github.com/gloo-foo/testable/run"

	command "github.com/gloo-foo/cmd-tail"
)

// errUpstream is a sentinel emitted by a deliberately failing source so byte
// mode's error-propagation path is observable via errors.Is.
var errUpstream = errors.New("upstream failed")

// failingSource adapts a Stream factory to gloo.Source[[]byte] for tests.
type failingSource func(context.Context) gloo.Stream[[]byte]

func (f failingSource) Stream(ctx context.Context) gloo.Stream[[]byte] { return f(ctx) }

// ==============================================================================
// Default Behavior (10 lines)
// ==============================================================================

func TestTail_DefaultTenLines(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 15; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "%d\n", i)
	}
	lines, err := testable.TestLines(command.Tail(), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"6", "7", "8", "9", "10", "11", "12", "13", "14", "15"})
}

func TestTail_LessThanDefault(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(), "1\n2\n3\n4\n5\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3", "4", "5"})
}

func TestTail_ExactlyTenLines(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 10; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "%d\n", i)
	}
	lines, err := testable.TestLines(command.Tail(), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
}

func TestTail_EmptyInput(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(), "")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

// ==============================================================================
// Custom Line Counts
// ==============================================================================

func TestTail_CustomThreeLines(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(command.TailLines(3)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"c", "d", "e"})
}

func TestTail_CustomOneLine(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(command.TailLines(1)), "first\nsecond\nthird\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"third"})
}

func TestTail_FewerLinesThanN(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(command.TailLines(100)), "1\n2\n3\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3"})
}

func TestTail_ExactlyN(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(command.TailLines(5)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c", "d", "e"})
}

// ==============================================================================
// Table-Driven
// ==============================================================================

func TestTail_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		n        command.TailLines
	}{
		{"three from five", "a\nb\nc\nd\ne\n", []string{"c", "d", "e"}, 3},
		{"one line", "first\nsecond\nthird\n", []string{"third"}, 1},
		{"all lines", "a\nb\n", []string{"a", "b"}, 5},
		{"with empty lines", "a\n\nb\nc\n", []string{"", "b", "c"}, 3},
		{"unicode", "hello\nworld\nend\n", []string{"world", "end"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := testable.TestLines(command.Tail(tt.n), run.Input(tt.input))
			assertion.NoError(t, err)
			assertion.Lines(t, lines, tt.expected)
		})
	}
}

// ==============================================================================
// Byte Count (-c)
// ==============================================================================

func TestTail_Bytes(t *testing.T) {
	// Input "hello world\n" → stream ["hello world"] → reconstituted "hello world\n" (12 bytes)
	// Last 6 bytes: "world\n" → TestLines trims trailing \n → "world"
	lines, err := testable.TestLines(command.Tail(command.TailBytes(6)), "hello world\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"world"})
}

func TestTail_Bytes_MoreThanInput(t *testing.T) {
	// Input "short\n" → stream ["short"] → reconstituted "short\n" (6 bytes)
	// N=100 exceeds length → all bytes: "short\n"
	// TestLines trims → "short"
	lines, err := testable.TestLines(command.Tail(command.TailBytes(100)), "short\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"short"})
}

func TestTail_Bytes_EmptyInput(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(command.TailBytes(5)), "")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

func TestTail_Bytes_MultipleLines(t *testing.T) {
	// Input "ab\ncd\n" → stream ["ab", "cd"] → reconstituted "ab\ncd\n" (6 bytes)
	// Last 4 bytes: "d\n" — wait: "ab\ncd\n"[2:] = "\ncd\n" (4 bytes)
	// TestLines: "\ncd\n" + "\n" = "\ncd\n\n" → trim → "\ncd" → split → ["", "cd"]
	lines, err := testable.TestLines(command.Tail(command.TailBytes(4)), "ab\ncd\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"", "cd"})
}

// ==============================================================================
// From Line (-n +N)
// ==============================================================================

func TestTail_FromLineStartsAtOffset(t *testing.T) {
	// GNU `tail -n +3` drops the first two lines and prints from line 3 onward.
	lines, err := testable.TestLines(command.Tail(command.TailFromLine(3)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"c", "d", "e"})
}

func TestTail_FromLineOneEmitsAll(t *testing.T) {
	// `tail -n +1` is the whole input — line 1 onward.
	lines, err := testable.TestLines(command.Tail(command.TailFromLine(1)), "a\nb\nc\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c"})
}

func TestTail_FromLineBeyondInput(t *testing.T) {
	// An offset past the last line yields no output.
	lines, err := testable.TestLines(command.Tail(command.TailFromLine(100)), "a\nb\nc\n")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

// ==============================================================================
// Non-positive Line Counts (default fallback)
// ==============================================================================

func TestTail_ZeroLinesFallsBackToDefault(t *testing.T) {
	// TailLines(0) is not "zero lines"; it resolves to the ten-line default.
	var b strings.Builder
	for i := 1; i <= 12; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "%d\n", i)
	}
	lines, err := testable.TestLines(command.Tail(command.TailLines(0)), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"3", "4", "5", "6", "7", "8", "9", "10", "11", "12"})
}

func TestTail_NegativeLinesFallsBackToDefault(t *testing.T) {
	// A negative count is likewise treated as the ten-line default.
	var b strings.Builder
	for i := 1; i <= 11; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "%d\n", i)
	}
	lines, err := testable.TestLines(command.Tail(command.TailLines(-5)), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "11"})
}

// TestTail_Bytes_PropagatesUpstreamError covers byte mode's error path: when
// the upstream stream fails, draining it surfaces the error, which Tail must
// forward downstream rather than swallow.
func TestTail_Bytes_PropagatesUpstreamError(t *testing.T) {
	src := failingSource(func(ctx context.Context) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
			if send([]byte("partial")) {
				sendErr(errUpstream)
			}
		})
	})

	_, err := gloo.From(context.Background(), src, command.Tail(command.TailBytes(3))).Collect()
	if !errors.Is(err, errUpstream) {
		t.Fatalf("got err %v, want %v", err, errUpstream)
	}
}

// ==============================================================================
// Edge Cases
// ==============================================================================

func TestTail_ManyLines(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 1000; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "line %d\n", i)
	}
	lines, err := testable.TestLines(command.Tail(command.TailLines(10)), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Count(t, lines, 10)
	assertion.Equal(t, lines[0], "line 991", "first line")
	assertion.Equal(t, lines[9], "line 1000", "last line")
}

func TestTail_SingleLine(t *testing.T) {
	lines, err := testable.TestLines(command.Tail(), "only\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"only"})
}
