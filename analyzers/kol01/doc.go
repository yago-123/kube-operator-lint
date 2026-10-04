// Package clientinreconcile will implement KOL01, which checks for Kubernetes
// clients created inside controller-runtime Reconcile methods. These clients
// should usually be created once and passed into the reconciler.
package clientinreconcile
