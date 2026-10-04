package clientinreconcile

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/cfg"
	"golang.org/x/tools/go/types/typeutil"
)

// checker owns the package-wide information reused while following calls from
// Reconcile methods into local helpers.
type checker struct {
	pass      *analysis.Pass
	functions map[*types.Func]*ast.FuncDecl
	graphs    map[*ast.FuncDecl]*functionGraph
}

type functionResult struct {
	createsClient bool
	// mayReturn tells the caller whether code after this function call can run.
	mayReturn bool
}

// functionGraph adds KOL01's conservative traversal boundary to Go's CFG.
// We currently stop at loops, switches, selects, and labels because following
// them confidently would require more state than this rule tracks.
type functionGraph struct {
	cfg     *cfg.CFG
	joins   map[*ast.IfStmt]*cfg.Block // If statement -> block where its branches meet.
	stopped map[ast.Node]bool
}

func newFunctionGraph(body *ast.BlockStmt, info *types.Info) *functionGraph {
	g := &functionGraph{
		cfg: cfg.New(body, func(call *ast.CallExpr) bool {
			// Built-in panic cannot return. Local helper calls are checked later,
			// when we know which functions are already on the call path.
			return typeutil.Callee(info, call) != types.Universe.Lookup("panic")
		}),
		joins:   make(map[*ast.IfStmt]*cfg.Block),
		stopped: make(map[ast.Node]bool),
	}
	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt,
			*ast.SelectStmt, *ast.LabeledStmt, *ast.BranchStmt:
			// CFG nodes may include a loop's initializer or a switch's tag.
			// Mark the whole construct to retain our existing analysis boundary.
			ast.Inspect(node, func(child ast.Node) bool {
				if child != nil {
					g.stopped[child] = true
				}
				return true
			})
			return false
		}
		return true
	})
	for _, block := range g.cfg.Blocks {
		if block.Kind == cfg.KindIfDone {
			if stmt, ok := block.Stmt.(*ast.IfStmt); ok {
				g.joins[stmt] = block
			}
		}
	}
	return g
}

// scanFunction follows one function's reachable CFG blocks. Reconcile scans
// report diagnostics directly. Helper scans only return a result, allowing the
// diagnostic to point at the helper call made by Reconcile.
func (c *checker) scanFunction(fn *types.Func, report bool, active map[*types.Func]bool) functionResult {
	if active[fn] {
		// We reached the same function on this call path. Stop the recursive
		// cycle without assuming execution continues after it.
		return functionResult{}
	}

	decl := c.functions[fn.Origin()]
	if decl == nil {
		return functionResult{mayReturn: true}
	}
	if c.graphs == nil {
		c.graphs = make(map[*ast.FuncDecl]*functionGraph)
	}
	graph := c.graphs[decl]
	if graph == nil {
		graph = newFunctionGraph(decl.Body, c.pass.TypesInfo)
		c.graphs[decl] = graph
	}

	active[fn] = true
	w := walker{
		checker: c, report: report, active: active, graph: graph,
		visited: make(map[blockState]bool), reported: make(map[token.Pos]bool),
	}
	w.block(graph.cfg.Blocks[0], false)
	delete(active, fn)
	return w.functionResult
}

// The same CFG block may be reached through cached and uncached paths. They
// must be visited separately or one path could silence the other by mistake.
type blockState struct {
	block      *cfg.Block
	reuseGuard bool
}

type walker struct {
	*checker
	functionResult
	report     bool
	reuseGuard bool
	active     map[*types.Func]bool
	graph      *functionGraph
	visited    map[blockState]bool
	reported   map[token.Pos]bool
}

// block visits each reachable CFG block once for each cache state. Go's CFG
// supplies normal sequencing and branch edges; KOL01 adds constant-condition,
// helper-call, and cached-client handling on top.
func (w *walker) block(block *cfg.Block, reuseGuard bool) {
	state := blockState{block, reuseGuard}
	if w.visited[state] {
		return
	}
	w.visited[state] = true
	w.reuseGuard = reuseGuard
	if w.graph.stopped[block.Stmt] {
		// Stopping our analysis does not mean the function stops running.
		w.mayReturn = true
		return
	}

	for i, node := range block.Nodes {
		if w.graph.stopped[node] {
			w.mayReturn = true
			return
		}
		if !w.node(node) {
			return
		}
		if _, ok := node.(*ast.ReturnStmt); ok {
			w.mayReturn = true
		}
		if stmt := sourceIf(block, i); stmt != nil && w.followIf(block, stmt) {
			return
		}
	}

	// Use the incoming cache state for every successor. A guard found while
	// visiting one branch must never leak into its sibling.
	guarded := w.reuseGuard
	for _, next := range block.Succs {
		w.block(next, guarded)
	}
}

// sourceIf identifies the condition node at the end of an if block. The true
// successor keeps the source statement, so no separate AST index is needed.
func sourceIf(block *cfg.Block, nodeIndex int) *ast.IfStmt {
	if nodeIndex != len(block.Nodes)-1 || len(block.Succs) != 2 || block.Succs[0].Kind != cfg.KindIfThen {
		return nil
	}
	stmt, _ := block.Succs[0].Stmt.(*ast.IfStmt)
	return stmt
}

// followIf handles branch decisions that need more context than the CFG has.
// It returns true when it has selected the next block itself.
func (w *walker) followIf(block *cfg.Block, stmt *ast.IfStmt) bool {
	// A CFG retains both edges even when the condition is a constant.
	if value, known := constantBool(w.pass.TypesInfo.Types[stmt.Cond].Value); known {
		edge := 0
		if !value {
			edge = 1
		}
		w.block(block.Succs[edge], w.reuseGuard)
		return true
	}

	if cached, reuse := w.isCachedClientGuard(stmt); cached {
		// Skip possible cache initialization and resume after the guard.
		w.block(w.graph.joins[stmt], w.reuseGuard || reuse)
		return true
	}

	if !w.report {
		// A helper branch depends on state supplied by its caller. Stop rather
		// than guessing which side runs and risking a false positive.
		w.mayReturn = true
		return true
	}
	return false
}

// node evaluates the expressions contained in one CFG node. It returns false
// when execution cannot continue, such as after a non-returning helper call.
func (w *walker) node(node ast.Node) bool {
	switch node := node.(type) {
	case ast.Expr:
		return w.expression(node)
	case *ast.ExprStmt:
		return w.expression(node.X)
	case *ast.AssignStmt:
		return w.expressions(node.Lhs) && w.expressions(node.Rhs)
	case *ast.ValueSpec:
		return w.expressions(node.Values)
	case *ast.ReturnStmt:
		return w.expressions(node.Results)
	case *ast.GoStmt:
		return w.expression(node.Call)
	case *ast.DeferStmt:
		return w.expression(node.Call)
	case *ast.IncDecStmt:
		return w.expression(node.X)
	case *ast.SendStmt:
		return w.expression(node.Chan) && w.expression(node.Value)
	default:
		return true
	}
}

func (w *walker) expressions(expressions []ast.Expr) bool {
	for _, expr := range expressions {
		if !w.expression(expr) {
			return false
		}
	}
	return true
}

// expression follows calls in Go's evaluation order. It deliberately ignores
// function-literal bodies and uncertain short-circuit operands: reporting only
// calls known to execute is safer than guessing and creating false positives.
func (w *walker) expression(expr ast.Expr) bool {
	canReturn := true
	ast.Inspect(expr, func(node ast.Node) bool {
		if !canReturn || node == nil {
			return false
		}

		currentExpr, isExpr := node.(ast.Expr)
		if isExpr {
			if value, ok := w.pass.TypesInfo.Types[currentExpr]; ok && value.Value != nil {
				// Constant expressions can contain unevaluated calls, for example
				// inside unsafe.Sizeof. Those calls do not create a client.
				return false
			}
		}

		switch node := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				canReturn = w.shortCircuit(node)
				return false
			}
		case *ast.CallExpr:
			canReturn = w.call(node)
			return false
		}
		return true
	})
	return canReturn
}

// shortCircuit visits the right operand only when the left operand proves it
// will run. An unknown value leaves the right side unchecked on purpose.
func (w *walker) shortCircuit(expr *ast.BinaryExpr) bool {
	if !w.expression(expr.X) {
		return false
	}
	left, known := constantBool(w.pass.TypesInfo.Types[expr.X].Value)
	if !known || (expr.Op == token.LAND && !left) || (expr.Op == token.LOR && left) {
		return true
	}
	return w.expression(expr.Y)
}

// call evaluates the function and arguments before inspecting the callee.
// Only statically resolved local functions are followed; interface dispatch
// and function variables are too ambiguous for this conservative rule.
func (w *walker) call(call *ast.CallExpr) bool {
	if !w.expression(call.Fun) || !w.expressions(call.Args) {
		return false
	}
	if builtin, ok := typeutil.Callee(w.pass.TypesInfo, call).(*types.Builtin); ok && builtin.Name() == "panic" {
		return false
	}

	callee := typeutil.StaticCallee(w.pass.TypesInfo, call)
	if callee == nil {
		return true
	}
	if isClientConstructor(callee) {
		if !w.reuseGuard {
			w.recordCreation(call)
		}
		return true
	}
	if callee.Pkg() != w.pass.Pkg {
		return true
	}

	result := w.scanFunction(callee, false, w.active)
	if result.createsClient {
		w.recordCreation(call)
	}
	return result.mayReturn
}

func (w *walker) recordCreation(call *ast.CallExpr) {
	w.createsClient = true
	// Branches may reach the same call with different cache-guard states.
	// Report the source location once, even if we visit it more than once.
	if w.report && !w.reported[call.Pos()] {
		w.reported[call.Pos()] = true
		w.pass.Reportf(call.Pos(), "%s", diagnostic)
	}
}

func constantBool(value constant.Value) (bool, bool) {
	if value == nil || value.Kind() != constant.Bool {
		return false, false
	}
	return constant.BoolVal(value), true
}
