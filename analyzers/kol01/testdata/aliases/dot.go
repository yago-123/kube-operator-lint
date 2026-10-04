// Checks that a dot-imported client-go constructor still reports, even though
// the call has no package qualifier.
package aliases

import (
	"context"

	. "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type DotImport struct{}

func (DotImport) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = NewForConfig(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}
