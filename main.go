// Command asl enforces the cocoonstack structural layout rules go vet cannot express.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/CMGS/asl/constraintname"
	"github.com/CMGS/asl/funcpartition"
	"github.com/CMGS/asl/functypedup"
	"github.com/CMGS/asl/methodpartition"
	"github.com/CMGS/asl/testorder"
	"github.com/CMGS/asl/topdecl"
)

func main() {
	multichecker.Main(
		constraintname.Analyzer,
		funcpartition.Analyzer,
		functypedup.Analyzer,
		methodpartition.Analyzer,
		testorder.Analyzer,
		topdecl.Analyzer,
	)
}
