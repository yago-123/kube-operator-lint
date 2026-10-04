package clientinreconcile

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/types/typeutil"
)

// isCachedClientGuard recognizes two simple lazy-initialization patterns:
//
//	if r.client != nil { return ... }
//	client, err := kubernetes.NewForConfig(...)
//
// and:
//
//	if r.client == nil { r.client, err = kubernetes.NewForConfig(...) }
//
// The first result says that the if statement is a cache guard. The second
// says that constructors after the guard belong to its initialization path.
func (w *walker) isCachedClientGuard(stmt *ast.IfStmt) (bool, bool) {
	field, op, ok := nilComparison(w.pass.TypesInfo, stmt.Cond)
	if !ok || !isKubernetesClientType(field.Type()) || stmt.Else != nil {
		return false, false
	}

	if op == token.NEQ && blockOnlyReturns(stmt.Body) {
		return true, true
	}
	if op == token.EQL && assignsConstructorToField(w.pass.TypesInfo, stmt.Body, field) {
		return true, false
	}
	return false, false
}

// nilComparison returns the selected field in "field == nil" or
// "field != nil". It uses type identity so same-named fields cannot match.
func nilComparison(info *types.Info, expr ast.Expr) (types.Object, token.Token, bool) {
	binary, ok := ast.Unparen(expr).(*ast.BinaryExpr)
	if !ok || (binary.Op != token.EQL && binary.Op != token.NEQ) {
		return nil, token.ILLEGAL, false
	}

	field, nilExpr := binary.X, binary.Y
	if ident, isIdent := ast.Unparen(field).(*ast.Ident); isIdent && ident.Name == "nil" {
		field, nilExpr = nilExpr, field
	}
	ident, ok := ast.Unparen(nilExpr).(*ast.Ident)
	if !ok || ident.Name != "nil" {
		return nil, token.ILLEGAL, false
	}

	selector, ok := ast.Unparen(field).(*ast.SelectorExpr)
	if !ok || info.Selections[selector] == nil {
		return nil, token.ILLEGAL, false
	}
	return info.Selections[selector].Obj(), binary.Op, true
}

func blockOnlyReturns(block *ast.BlockStmt) bool {
	if len(block.List) != 1 {
		return false
	}
	_, ok := block.List[0].(*ast.ReturnStmt)
	return ok
}

func assignsConstructorToField(info *types.Info, block *ast.BlockStmt, field types.Object) bool {
	for _, stmt := range block.List {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range assign.Lhs {
			selector, isSelector := ast.Unparen(lhs).(*ast.SelectorExpr)
			if !isSelector || info.Selections[selector] == nil || info.Selections[selector].Obj() != field {
				continue
			}
			for _, rhs := range assign.Rhs {
				if containsClientConstructor(info, rhs) {
					return true
				}
			}
		}
	}
	return false
}

func containsClientConstructor(info *types.Info, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok {
			callee := typeutil.StaticCallee(info, call)
			found = callee != nil && isClientConstructor(callee)
		}
		return !found
	})
	return found
}

func isKubernetesClientType(t types.Type) bool {
	t = types.Unalias(t)
	if pointer, ok := t.(*types.Pointer); ok {
		t = types.Unalias(pointer.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	switch named.Obj().Pkg().Path() {
	case "k8s.io/client-go/kubernetes":
		return named.Obj().Name() == "Clientset"
	case "k8s.io/client-go/dynamic":
		return named.Obj().Name() == "DynamicClient"
	case "sigs.k8s.io/controller-runtime/pkg/client":
		return named.Obj().Name() == "Client"
	default:
		return false
	}
}
