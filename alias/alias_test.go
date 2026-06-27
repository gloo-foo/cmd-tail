package alias_test

import (
	"slices"
	"testing"

	tail "github.com/gloo-foo/cmd-tail/alias"
	"github.com/gloo-foo/testable"
)

// The alias package re-exports the tail constructor and its flag types under
// unprefixed names. A mis-wired re-export (say, Lines aliased to TailBytes, or
// Tail bound to the wrong function) compiles cleanly, so only behavior can
// prove the wiring. Each test exercises one re-export and asserts the GNU tail
// output it must produce.

const numbered = "1\n2\n3\n4\n5\n"

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAlias_DefaultEmitsLastTen(t *testing.T) {
	// No flag: the last ten lines. With five lines in, all five come out.
	lines, err := testable.TestLines(tail.Tail(), numbered)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"1", "2", "3", "4", "5"})
}

func TestAlias_LinesSelectsLastN(t *testing.T) {
	// -n 2 keeps the final two lines.
	lines, err := testable.TestLines(tail.Tail(tail.Lines(2)), numbered)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"4", "5"})
}

func TestAlias_FromLineStartsAtOffset(t *testing.T) {
	// -n +3 drops the first two lines and emits from line three onward.
	lines, err := testable.TestLines(tail.Tail(tail.FromLine(3)), numbered)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"3", "4", "5"})
}

func TestAlias_BytesSelectsLastN(t *testing.T) {
	// -c 2 keeps the final two bytes: "5\n", which TestLines trims to "5".
	lines, err := testable.TestLines(tail.Tail(tail.Bytes(2)), numbered)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"5"})
}
