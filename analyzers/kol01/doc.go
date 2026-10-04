// Ensure Kubernetes clients are initialized once during setup and reused in `Reconcile`.
//
// ```go
// // Avoid: creating a client inside Reconcile.
//
//	func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//		c, err := client.New(r.Config, client.Options{}) // KOL01
//		if err != nil {
//			return ctrl.Result{}, err
//		}
//
//		var pod corev1.Pod
//		err = c.Get(ctx, req.NamespacedName, &pod)
//		return ctrl.Result{}, client.IgnoreNotFound(err)
//	}
//
// // Prefer: reusing the client injected during setup.
//
//	func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//		var pod corev1.Pod
//		err := r.Client.Get(ctx, req.NamespacedName, &pod)
//		return ctrl.Result{}, client.IgnoreNotFound(err)
//	}
//
// ```
//
// Reconcile runs repeatedly. Creating a fresh Kubernetes client on each call
// repeats initialization work and makes it harder to inject a fake in tests.
// Diagnostics start with "kol01-clientinreconcile" and point to the call that
// triggers client creation.
//
// # How to fix
//
// Supply the client during controller setup, then use `r.Client` in `Reconcile`.
// For controller-runtime, use the manager's existing client:
//
// ```go
//
//	type Reconciler struct {
//		Client client.Client
//	}
//
// // During controller setup:
// r := &Reconciler{Client: mgr.GetClient()}
// ```
//
// For client-go's kubernetes or dynamic clients, call `NewForConfig` once during
// setup, handle its error there, and pass the resulting client to the reconciler.
//
// # Other cases covered
//
// Beyond direct calls, `kol01` also covers:
//
//   - Local helper chains, including helpers in other files of the same package.
//   - Concrete methods, including methods promoted from embedded types.
//   - Constructors in evaluated if initializers, conditions, and branches.
//   - Import aliases, dot imports, and type aliases.
//   - Reconcile methods in test files.
//
// For a helper, the diagnostic points to the call inside `Reconcile`:
//
// ```go
// // Inside Reconcile:
// c, err := newClient(r.Config) // KOL01
//
//	func newClient(cfg *rest.Config) (client.Client, error) {
//		return client.New(cfg, client.Options{})
//	}
//
// ```
//
// Follow the helper's calls to the constructor and move client creation to setup.
// In helpers, conditional branches are skipped unless a nil check clearly guards
// a cached client. Loop and switch bodies are also left alone.
//
// Supported constructors:
//
// - `k8s.io/client-go/kubernetes`: `NewForConfig` and `NewForConfigOrDie`.
// - `k8s.io/client-go/dynamic`: `NewForConfig` and `NewForConfigOrDie`.
// - `sigs.k8s.io/controller-runtime/pkg/client`: `New`.
//
// Unrelated APIs with matching names, fake-client builders, and reuse of
// injected or cached clients do not trigger kol01.
package clientinreconcile
