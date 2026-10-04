// Package analyzers keeps the CLI and golangci-lint using the same set of rules.
package analyzers

import "golang.org/x/tools/go/analysis"

// All returns the available rules. Add new rules here so both entry points pick
// them up.
func All() []*analysis.Analyzer {
	// todo: this stays empty until the first rule is implemented and tested.
	return []*analysis.Analyzer{}
}
