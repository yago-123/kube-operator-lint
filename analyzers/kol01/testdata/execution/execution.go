// Checks that traversal follows expression evaluation and unconditional blocks.
// Constructors in declarations, arguments, receivers, and returns should report.
// Short-circuited or unevaluated expressions, closures, and code after a return
// or panic should stay quiet. Deferred and asynchronous calls still run helpers.
package execution

import (
	"context"
	"unsafe"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Expressions struct{}

func (Expressions) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	{
		var c = kubernetes.NewForConfigOrDie(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
		_ = c
	}
	consume(kubernetes.NewForConfigOrDie(&rest.Config{}))     // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_ = kubernetes.NewForConfigOrDie(&rest.Config{}).CoreV1() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return resultWithClient()                                 // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
}

func consume(*kubernetes.Clientset) {}

func resultWithClient() (ctrl.Result, error) {
	_, err := kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, err
}

type IfCondition struct{}

func (IfCondition) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	if createsClient() { // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

type ShortCircuit struct{}

func (ShortCircuit) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = false && createsClient()
	_ = true || createsClient()
	_ = createsClient() && false // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

func createsClient() bool {
	_ = kubernetes.NewForConfigOrDie(&rest.Config{})
	return true
}

type Unevaluated struct{}

func (Unevaluated) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = unsafe.Sizeof(kubernetes.NewForConfigOrDie(&rest.Config{}))
	_ = unsafe.Alignof(kubernetes.NewForConfigOrDie(&rest.Config{}))
	var unused = func() { _, _ = kubernetes.NewForConfig(&rest.Config{}) }
	_ = unused
	return ctrl.Result{}, nil
}

type ConstantBranch struct{}

func (ConstantBranch) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	if false {
		_, _ = kubernetes.NewForConfig(&rest.Config{})
	}
	return ctrl.Result{}, nil
}

type Conditional struct{}

func (Conditional) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace != "" {
		_, _ = kubernetes.NewForConfig(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	}
	return ctrl.Result{}, nil
}

type AfterReturn struct{}

func (AfterReturn) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	{
		return ctrl.Result{}, nil
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type AfterPanic struct{}

func (AfterPanic) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	fail()
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

func fail() {
	panic("stop")
	_, _ = kubernetes.NewForConfig(&rest.Config{})
}

type PanickingArgument struct{}

func (PanickingArgument) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = kubernetes.NewForConfigOrDie(failedConfig())
	return ctrl.Result{}, nil
}

func failedConfig() *rest.Config {
	panic("no configuration")
}

type Deferred struct{}

func (Deferred) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	defer createsClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type Asynchronous struct{}

func (Asynchronous) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	go createsClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type RecursiveArgument struct{}

func (RecursiveArgument) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = kubernetes.NewForConfigOrDie(recursiveConfig())
	return ctrl.Result{}, nil
}

func recursiveConfig() *rest.Config {
	return recursiveConfig()
}
