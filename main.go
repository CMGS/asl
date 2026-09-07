// Command asl enforces the cocoonstack structural layout rules go vet cannot express.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/CMGS/asl/cmpor"
	"github.com/CMGS/asl/constraintname"
	"github.com/CMGS/asl/forwarder"
	"github.com/CMGS/asl/funcpartition"
	"github.com/CMGS/asl/functypedup"
	"github.com/CMGS/asl/labelenum"
	"github.com/CMGS/asl/methodinterleave"
	"github.com/CMGS/asl/methodpartition"
	"github.com/CMGS/asl/testorder"
	"github.com/CMGS/asl/topdecl"
	"github.com/CMGS/asl/typeblockgap"
)

func main() {
	multichecker.Main(
		cmpor.Analyzer,
		constraintname.Analyzer,
		forwarder.Analyzer,
		funcpartition.Analyzer,
		functypedup.Analyzer,
		labelenum.Analyzer,
		methodinterleave.Analyzer,
		methodpartition.Analyzer,
		testorder.Analyzer,
		topdecl.Analyzer,
		typeblockgap.Analyzer,
	)
}
