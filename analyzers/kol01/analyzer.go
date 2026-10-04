package clientinreconcile

import "golang.org/x/tools/go/analysis"

// Analyzer checks KOL01: Kubernetes client construction during reconciliation.
var Analyzer = &analysis.Analyzer{
	Name: "clientinreconcile",
	Doc:  "check for Kubernetes clients created inside Reconcile or its local helpers",
	Run:  run,
}

func run(*analysis.Pass) (any, error) {
	// This stub lets the tests fail on missing diagnostics while we review the
	// rule's behavior. Detection logic comes after the test specification.
	return nil, nil
}
