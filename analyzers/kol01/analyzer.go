package clientinreconcile

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Keep the rule name in the message so diagnostics stay easy to identify in any runner.
const diagnostic = "kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler"

// Analyzer checks KOL01: Kubernetes client construction during reconciliation.
var Analyzer = &analysis.Analyzer{
	Name: "clientinreconcile",
	Doc:  "check for Kubernetes clients created inside Reconcile or its local helpers",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	functions := make(map[*types.Func]*ast.FuncDecl)
	var ordered []*types.Func
	// Index bodies by their resolved function objects so helper calls can be
	// followed by identity rather than by the spelling of their names.
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			object, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
			if ok {
				object = object.Origin()
				functions[object] = fn
				ordered = append(ordered, object)
			}
		}
	}

	checker := &checker{pass: pass, functions: functions}
	// Start only at methods with the controller-runtime signature; helpers are
	// visited from these roots, not analyzed as reconciliations themselves.
	for _, fn := range ordered {
		if isControllerRuntimeReconcile(fn) {
			checker.scanFunction(fn, true, make(map[*types.Func]bool))
		}
	}
	return nil, nil
}

// Reconcile's signature is a reliable signal that this method belongs to a
// controller-runtime reconciler, even when its imports or types are aliased.
func isControllerRuntimeReconcile(fn *types.Func) bool {
	if fn.Name() != "Reconcile" || fn.Signature().Recv() == nil {
		return false
	}

	sig := fn.Signature()
	params, results := sig.Params(), sig.Results()
	return params.Len() == 2 &&
		namedType(params.At(0).Type(), "context", "Context") &&
		namedType(params.At(1).Type(), "sigs.k8s.io/controller-runtime/pkg/reconcile", "Request") &&
		results.Len() == 2 &&
		namedType(results.At(0).Type(), "sigs.k8s.io/controller-runtime/pkg/reconcile", "Result") &&
		types.Identical(types.Unalias(results.At(1).Type()), types.Universe.Lookup("error").Type())
}

// Package paths and type names stay the same when source imports use aliases.
func namedType(t types.Type, packagePath, name string) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	object := named.Obj()
	return object.Name() == name && object.Pkg() != nil && object.Pkg().Path() == packagePath
}

// These are package functions. Checking the package scope excludes methods
// that happen to use the same constructor names.
func isClientConstructor(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil || pkg.Scope().Lookup(fn.Name()) != fn {
		return false
	}

	switch pkg.Path() {
	case "k8s.io/client-go/kubernetes", "k8s.io/client-go/dynamic":
		return fn.Name() == "NewForConfig" || fn.Name() == "NewForConfigOrDie"
	case "sigs.k8s.io/controller-runtime/pkg/client":
		return fn.Name() == "New"
	default:
		return false
	}
}
