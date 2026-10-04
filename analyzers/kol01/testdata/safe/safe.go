// Checks that setup-time construction, injected or cached clients, fake
// clients, and unreachable constructors stay quiet. Initialization guarded
// by a cache check or sync.Once should also stay quiet, both in Reconcile
// and through helpers.
package safe

import (
	"context"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var startupClient = kubernetes.NewForConfigOrDie(&rest.Config{})

func setup() (client.Client, error) {
	return client.New(&rest.Config{}, client.Options{})
}

type Injected struct {
	Client client.Client
}

func (r *Injected) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = r.Client
	_ = startupClient
	return ctrl.Result{}, nil
}

type Embedded struct {
	client.Client
}

func (r *Embedded) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = r.Client
	return ctrl.Result{}, nil
}

type FakeClient struct{}

func (FakeClient) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = fake.NewClientBuilder().Build()
	_ = buildFake()
	return ctrl.Result{}, nil
}

func buildFake() client.Client {
	return fake.NewClientBuilder().Build()
}

type Cached struct {
	client *kubernetes.Clientset
}

func (r *Cached) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = r.existingClient()
	_, _ = r.lazyClient()
	_ = r.lazyClientInBranch()
	_, _ = r.throughAnotherHelper()
	return ctrl.Result{}, nil
}

func (r *Cached) existingClient() *kubernetes.Clientset {
	return r.client
}

func (r *Cached) lazyClient() (*kubernetes.Clientset, error) {
	if r.client != nil {
		return r.client, nil
	}
	c, err := kubernetes.NewForConfig(&rest.Config{})
	if err != nil {
		return nil, err
	}
	r.client = c
	return c, nil
}

func (r *Cached) lazyClientInBranch() *kubernetes.Clientset {
	if r.client == nil {
		r.client = kubernetes.NewForConfigOrDie(&rest.Config{})
	}
	return r.client
}

func (r *Cached) throughAnotherHelper() (*kubernetes.Clientset, error) {
	return r.lazyClient()
}

type Once struct {
	once   sync.Once
	client *kubernetes.Clientset
}

func (r *Once) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	_ = r.sharedClient()
	return ctrl.Result{}, nil
}

func (r *Once) sharedClient() *kubernetes.Clientset {
	r.once.Do(func() {
		r.client = kubernetes.NewForConfigOrDie(&rest.Config{})
	})
	return r.client
}

type DirectCache struct {
	client *kubernetes.Clientset
}

func (r *DirectCache) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	if r.client == nil {
		r.client = kubernetes.NewForConfigOrDie(&rest.Config{})
	}
	return ctrl.Result{}, nil
}

type DirectOnce struct {
	once sync.Once
}

func (r *DirectOnce) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	r.once.Do(func() {
		_, _ = kubernetes.NewForConfig(&rest.Config{})
	})
	return ctrl.Result{}, nil
}

type Unreachable struct{}

func (Unreachable) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	nothingToDo()
	return ctrl.Result{}, nil
}

func nothingToDo() {
	return
	_, _ = kubernetes.NewForConfig(&rest.Config{})
}
