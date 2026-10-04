// Checks client creation through local helper chains and concrete methods,
// including embedded methods and helpers in another file. Each call in
// Reconcile should report once, even when a helper creates multiple clients
// or is shared by several calls or controllers.
package helpers

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct {
	factory *Factory
}

func (r *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_, _ = newClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = firstHop() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = r.factory.NewClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_ = r.factory.NewDynamicClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = (*Factory).NewClient(r.factory) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = r.newRuntimeClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	createTwoClients() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = newClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type AnotherController struct{}

func (AnotherController) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	if _, err := newClient(); err != nil { // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

type EmbeddedFactory struct {
	Factory
}

func (r *EmbeddedFactory) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = r.NewClient() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}
