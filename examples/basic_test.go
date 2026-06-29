package tail_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-tail"
)

func ExampleTail_basic() {
	// echo "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12" | tail
	output, _ := testable.Test(
		command.Tail(),
		"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12",
	)
	fmt.Print(output)
	// Output:
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 10
	// 11
	// 12
}
