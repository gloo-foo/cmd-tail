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

// flags is the folded option set for a tail run. The zero value selects the
// default behavior: the last ten lines.
type flags struct {
	lines    TailLines
	fromLine TailFromLine
	bytes    TailBytes
}

// with folds one option value into the flag set. Values of any other type are
// ignored: tail reads from the pipeline and takes no positional arguments.
func (f flags) with(o any) flags {
	switch v := o.(type) {
	case TailLines:
		f.lines = v
	case TailFromLine:
		f.fromLine = v
	case TailBytes:
		f.bytes = v
	}
	return f
}

// fold collapses the Tail option values into the flag set.
func fold(opts []any) flags {
	var f flags
	for _, o := range opts {
		f = f.with(o)
	}
	return f
}
