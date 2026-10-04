// Package kubeoperatorlint makes the rules available as a golangci-lint module
// plugin.
package kubeoperatorlint

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/yago-123/kube-operator-lint/analyzers"
)

func init() {
	// golangci-lint imports this package to register the plugin.
	register.Plugin("kube-operator-lint", newPlugin)
}

func newPlugin(_ any) (register.LinterPlugin, error) {
	// todo: we don't have configurable rules yet, so there's nothing to read from settings.
	return &Plugin{}, nil
}

// Plugin connects golangci-lint to the analyzer list. The rule logic stays in
// the analyzer packages so it can also run without golangci-lint.
type Plugin struct{}

// BuildAnalyzers returns the same rules used by the standalone command.
func (*Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return analyzers.All(), nil
}

// GetLoadMode tells golangci-lint which package data the rules need.
func (*Plugin) GetLoadMode() string {
	// Type information lets us distinguish Kubernetes APIs from unrelated code
	// with the same names.
	return register.LoadModeTypesInfo
}
