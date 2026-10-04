// Checks that familiar names alone do not trigger KOL01. Unrelated packages,
// local functions, mocks, shadowed imports, and methods without the exact
// controller-runtime Reconcile signature should all stay quiet.
package lookalikes

import (
	"context"

	kubernetes "example.com/kol01fixtures/lookalike/kubernetes"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct{}

func (Controller) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = kubernetes.NewForConfig(&rest.Config{})
	_ = kubernetes.NewForConfigOrDie(&rest.Config{})

	_, _ = NewForConfig(&rest.Config{})

	_, _ = MockFactory{}.NewForConfig(&rest.Config{})
	_, _ = LocalFactory{}.New(&rest.Config{})
	return ctrl.Result{}, nil
}

func NewForConfig(*rest.Config) (*kube.Clientset, error) {
	return &kube.Clientset{}, nil
}

type MockFactory struct{}

func (MockFactory) NewForConfig(*rest.Config) (*kube.Clientset, error) {
	return &kube.Clientset{}, nil
}

type LocalFactory struct{}

func (LocalFactory) New(*rest.Config) (*kube.Clientset, error) {
	return &kube.Clientset{}, nil
}

type ShadowedImport struct{}

func (ShadowedImport) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	kube := MockFactory{}
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type Request ctrl.Request
type Result ctrl.Result
type Context interface{ context.Context }
type Error interface{ Error() string }

type WrongRequest struct{}

func (WrongRequest) Reconcile(context.Context, Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type WrongResult struct{}

func (WrongResult) Reconcile(context.Context, ctrl.Request) (Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return Result{}, nil
}

type WrongContext struct{}

func (WrongContext) Reconcile(Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type WrongError struct{}

func (WrongError) Reconcile(context.Context, ctrl.Request) (ctrl.Result, Error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type PointerRequest struct{}

func (PointerRequest) Reconcile(context.Context, *ctrl.Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type ExtraParameter struct{}

func (ExtraParameter) Reconcile(context.Context, ctrl.Request, bool) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type VariadicRequest struct{}

func (VariadicRequest) Reconcile(context.Context, ...ctrl.Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type DifferentName struct{}

func (DifferentName) Sync(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

func Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_, _ = kube.NewForConfig(&rest.Config{})
	return ctrl.Result{}, nil
}

type FakeReconciler struct{}

func (FakeReconciler) Reconcile(context.Context, string) error {
	_, err := kube.NewForConfig(&rest.Config{})
	return err
}
