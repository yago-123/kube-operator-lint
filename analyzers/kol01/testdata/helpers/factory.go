// Provides the helpers called from reconcile.go to check traversal across
// files. Helper bodies and setup calls should stay quiet; diagnostics belong
// to the calls that trigger construction inside Reconcile.
package helpers

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(&rest.Config{})
}

func firstHop() (*kubernetes.Clientset, error) {
	return secondHop()
}

func secondHop() (*kubernetes.Clientset, error) {
	return newClient()
}

type Factory struct{}

func (*Factory) NewClient() (*kubernetes.Clientset, error) {
	c, err := newClient()
	return c, err
}

func (Factory) NewDynamicClient() *dynamic.DynamicClient {
	return dynamic.NewForConfigOrDie(&rest.Config{})
}

func (*Controller) newRuntimeClient() (client.Client, error) {
	return client.New(&rest.Config{}, client.Options{})
}

func createTwoClients() {
	_, _ = newClient()
	_, _ = dynamic.NewForConfig(&rest.Config{})
}

func setup() (*kubernetes.Clientset, error) {
	return newClient()
}
