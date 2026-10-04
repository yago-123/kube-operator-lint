// Checks that all five supported client constructors report when called directly
// inside a controller-runtime Reconcile, including if initializers and methods
// with unnamed parameters or named results.
package direct

import (
	"context"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Controller struct{}

var _ reconcile.Reconciler = (*Controller)(nil)

func (*Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cfg := &rest.Config{}
	_, _ = kubernetes.NewForConfig(cfg)      // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_ = kubernetes.NewForConfigOrDie(cfg)    // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_, _ = dynamic.NewForConfig(cfg)         // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_ = dynamic.NewForConfigOrDie(cfg)       // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_, _ = client.New(cfg, client.Options{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type IfInitializer struct{}

func (IfInitializer) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if _, err := kubernetes.NewForConfig(&rest.Config{}); err != nil { // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

type NamedResults struct{}

func (NamedResults) Reconcile(context.Context, reconcile.Request) (result reconcile.Result, err error) {
	_, err = client.New(&rest.Config{}, client.Options{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return
}
