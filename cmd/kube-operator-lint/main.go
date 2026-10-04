package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/yago-123/kube-operator-lint/analyzers"
)

func main() {
	// Let multichecker handle flags and package loading for the shared rules.
	multichecker.Main(analyzers.All()...)
}
