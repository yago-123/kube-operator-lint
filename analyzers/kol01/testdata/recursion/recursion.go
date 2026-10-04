// Checks that recursive helper traversal terminates. Client creation before
// a cycle should report; cycles without construction and constructors after
// calls that never return should stay quiet.
package recursion

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct{}

func (Controller) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	createBeforeCycle() // want "^kol01-clientinreconcile: Kubernetes client created inside Reconcile; initialize the client once and inject it into the reconciler$"
	return ctrl.Result{}, nil
}

func createBeforeCycle() {
	_ = kubernetes.NewForConfigOrDie(&rest.Config{})
	cycleA()
}

func cycleA() { cycleB() }
func cycleB() { cycleA() }

type PureCycle struct{}

func (PureCycle) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	cycleA()
	return ctrl.Result{}, nil
}

type SelfRecursion struct{}

func (SelfRecursion) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	self()
	return ctrl.Result{}, nil
}

func self() { self() }

type UnreachableCreation struct{}

func (UnreachableCreation) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	afterCycle()
	return ctrl.Result{}, nil
}

func afterCycle() {
	self()
	_ = kubernetes.NewForConfigOrDie(&rest.Config{})
}
