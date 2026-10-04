// Checks that import aliases, chained type aliases, and parentheses around
// a constructor do not hide its identity. Each real constructor should report.
package aliases

import (
	ctxpkg "context"

	dyn "k8s.io/client-go/dynamic"
	bananas "k8s.io/client-go/kubernetes"
	config "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"
	runtime "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Context = ctxpkg.Context
type Request = runtime.Request
type Result = controllerruntime.Result
type AnotherRequest = Request

type Controller struct{}

func (Controller) Reconcile(c Context, key AnotherRequest) (Result, error) {
	_, _ = bananas.NewForConfig(&config.Config{})                 // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_ = bananas.NewForConfigOrDie(&config.Config{})               // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_, _ = dyn.NewForConfig(&config.Config{})                     // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_ = dyn.NewForConfigOrDie(&config.Config{})                   // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	_, _ = kubeclient.New(&config.Config{}, kubeclient.Options{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"

	_, _ = (bananas.NewForConfig)(&config.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return Result{}, nil
}
