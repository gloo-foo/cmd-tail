package command

// TailLines sets the number of trailing lines to output (-n N). Default: 10.
type TailLines int

// TailFromLine outputs every line from line N onward, 1-indexed (-n +N).
// TailFromLine(1) emits the whole input; TailFromLine(3) drops the first two
// lines. A value of zero or below leaves the from-line mode disabled.
type TailFromLine int

// TailBytes outputs the last N bytes instead of lines (-c N). When set, the
// command reconstitutes the input stream and emits the last N bytes.
type TailBytes int

// flags is the parsed flag state for a tail run. The zero value selects the
// default behavior: the last ten lines.
type flags struct {
	lines    TailLines
	fromLine TailFromLine
	bytes    TailBytes
}

func (l TailLines) Configure(f *flags)    { f.lines = l }
func (n TailFromLine) Configure(f *flags) { f.fromLine = n }
func (b TailBytes) Configure(f *flags)    { f.bytes = b }
