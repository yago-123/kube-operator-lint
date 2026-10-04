package clientinreconcile_test

import (
	"os/exec"
	"testing"

	clientinreconcile "github.com/yago-123/kube-operator-lint/analyzers/kol01"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestKOL01ClientInReconcile(t *testing.T) {
	testdata := analysistest.TestData()

	// analysistest disables module downloads while loading fixtures. Fetch the
	// fixture module's real dependencies first, including on a fresh checkout.
	download := exec.CommandContext(t.Context(), "go", "mod", "download")
	download.Dir = testdata
	if output, err := download.CombinedOutput(); err != nil {
		t.Fatalf("download fixture dependencies: %v\n%s", err, output)
	}

	// Each fixture package covers one part of the rule. Lines without a want
	// comment must stay quiet, including the constructors inside helper bodies:
	// callers in Reconcile should get the diagnostic, not the shared helper.
	analysistest.Run(t, testdata, clientinreconcile.Analyzer, "./...")
}
