// Checks that branches rejoin without duplicate diagnostics and that returns
// and constant conditions keep unreachable constructors quiet. Loops, switches,
// selects, and labels remain outside the rule's current traversal scope.
package execution

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type JoinedBranches struct{}

func (JoinedBranches) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace == "" {
		_ = req.Name
	} else {
		_ = req.Namespace
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type BothReturn struct{}

func (BothReturn) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace == "" {
		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{}, nil
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type OneReturns struct{}

func (OneReturns) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace == "" {
		return ctrl.Result{}, nil
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type ConstantElse struct{}

func (ConstantElse) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	if true {
		return ctrl.Result{}, nil
	} else {
		_, _ = kubernetes.NewForConfig(&rest.Config{})
	}
	return ctrl.Result{}, nil
}

type Loop struct{}

func (Loop) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	for createsClient() {
		_, _ = kubernetes.NewForConfig(&rest.Config{})
		break
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type Switch struct{}

func (Switch) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	switch createsClient() {
	case true:
		_, _ = kubernetes.NewForConfig(&rest.Config{})
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type Select struct{}

func (Select) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	select {
	default:
		_, _ = kubernetes.NewForConfig(&rest.Config{})
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type Label struct{}

func (Label) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	goto construct
construct:
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}
