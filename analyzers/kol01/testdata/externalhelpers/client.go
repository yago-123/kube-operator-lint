// Provides a real constructor behind a helper in another package. Calls from
// unsupported should stay quiet because traversal only follows local helpers;
// the helper definition itself should not produce a diagnostic.
package externalhelpers

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(&rest.Config{})
}
