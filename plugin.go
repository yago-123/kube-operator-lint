// Package kubeoperatorlint makes the rules available as a golangci-lint module
// plugin.
package kubeoperatorlint

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	clientinreconcile "github.com/yago-123/kube-operator-lint/analyzers/kol01"
)

func init() {
	// Register each rule separately so projects can enable and suppress KOLs
	// using golangci-lint's normal per-linter controls.
	register.Plugin("kol01", pluginFor(clientinreconcile.Analyzer))
}

// pluginFor keeps future registrations small: each KOL gets its own name and
// analyzer while sharing the thin golangci-lint adapter.
func pluginFor(analyzer *analysis.Analyzer) register.NewPlugin {
	return func(_ any) (register.LinterPlugin, error) {
		return &Plugin{analyzers: []*analysis.Analyzer{analyzer}}, nil
	}
}

// Plugin exposes one KOL rule through golangci-lint. The rule logic stays in
// its analyzer package so the standalone command can still run every rule.
type Plugin struct {
	analyzers []*analysis.Analyzer
}

// BuildAnalyzers returns the analyzer assigned to this golangci-lint entry.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return p.analyzers, nil
}

// GetLoadMode tells golangci-lint which package data the rules need.
func (*Plugin) GetLoadMode() string {
	// Type information lets us distinguish Kubernetes APIs from unrelated code
	// with the same names.
	return register.LoadModeTypesInfo
}
