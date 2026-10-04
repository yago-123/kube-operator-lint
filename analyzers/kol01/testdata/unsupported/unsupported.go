// Records cases deliberately outside the rule's scope: interface dispatch,
// function and method values, external helpers, callbacks, and helper guards
// that require argument propagation. All should stay quiet; these are
// analysis limits, not recommended controller patterns.
package unsupported

import (
	"context"

	"example.com/kol01fixtures/externalhelpers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Builder interface {
	NewClient() (*kubernetes.Clientset, error)
}

type Controller struct {
	builder Builder
	factory func() (*kubernetes.Clientset, error)
}

func (r *Controller) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = r.builder.NewClient()

	_, _ = r.factory()

	_, _ = externalhelpers.NewClient()

	constructor := kubernetes.NewForConfig
	_, _ = constructor(&rest.Config{})

	local := Factory{}
	methodValue := local.NewClient
	_, _ = methodValue()
	return ctrl.Result{}, nil
}

type Factory struct{}

func (Factory) NewClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(&rest.Config{})
}

type Closures struct{}

func (Closures) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	callback := func() { _, _ = kubernetes.NewForConfig(&rest.Config{}) }
	_ = callback
	acceptCallback(func() { _, _ = kubernetes.NewForConfig(&rest.Config{}) })
	return ctrl.Result{}, nil
}

func acceptCallback(func()) {}

type Guarded struct{}

func (Guarded) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	guardedHelper(true)
	return ctrl.Result{}, nil
}

func guardedHelper(enabled bool) {
	if !enabled {
		return
	}
	_, _ = kubernetes.NewForConfig(&rest.Config{})
}
