// Package clientinreconcile will check for Kubernetes clients created inside
// controller-runtime Reconcile methods. These clients should usually be created
// once and passed into the reconciler.
package clientinreconcile
