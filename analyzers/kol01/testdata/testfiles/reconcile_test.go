// Checks that test files are analyzed too: real client construction inside
// Reconcile should report, while a fake client builder should stay quiet.
package testfiles

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type Controller struct{}

func (Controller) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = kubernetes.NewForConfig(&rest.Config{}) // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

type FakeController struct{}

func (FakeController) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = fake.NewClientBuilder().Build()
	return ctrl.Result{}, nil
}
