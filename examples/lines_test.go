package tail_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-tail"
)

func ExampleTail_lines() {
	// echo "1\n2\n3\n4\n5" | tail -n 3
	output, _ := testable.Test(
		command.Tail(command.TailLines(3)),
		"1\n2\n3\n4\n5",
	)
	fmt.Print(output)
	// Output:
	// 3
	// 4
	// 5
}
