// Package analyzers keeps the CLI and golangci-lint using the same set of rules.
package analyzers

import (
	"golang.org/x/tools/go/analysis"

	kol01 "github.com/yago-123/kube-operator-lint/analyzers/kol01"
)

// All returns the available rules. Add new rules here so both entry points pick
// them up.
func All() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		kol01.Analyzer,
	}
}
