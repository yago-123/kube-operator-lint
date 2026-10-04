package kubeoperatorlint_test

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"

	_ "github.com/yago-123/kube-operator-lint"
	clientinreconcile "github.com/yago-123/kube-operator-lint/analyzers/kol01"
)

func TestKOL01Plugin(t *testing.T) {
	t.Parallel()

	factory, err := register.GetPlugin("kol01")
	if err != nil {
		t.Fatalf("find kol01 plugin: %v", err)
	}
	plugin, err := factory(nil)
	if err != nil {
		t.Fatalf("create kol01 plugin: %v", err)
	}

	analyzers, err := plugin.BuildAnalyzers()
	if err != nil {
		t.Fatalf("build kol01 analyzers: %v", err)
	}
	if len(analyzers) != 1 || analyzers[0] != clientinreconcile.Analyzer {
		t.Fatalf("kol01 returned unexpected analyzers: %v", analyzers)
	}
	if plugin.GetLoadMode() != register.LoadModeTypesInfo {
		t.Fatalf("kol01 load mode = %q, want %q", plugin.GetLoadMode(), register.LoadModeTypesInfo)
	}
}
