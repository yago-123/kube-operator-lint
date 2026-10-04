// Provides a package with the same name and constructor names as client-go,
// but a different import path. Calls to these lookalikes should stay quiet.
package kubernetes

import "k8s.io/client-go/rest"

type Client struct{}

func NewForConfig(*rest.Config) (*Client, error) { return &Client{}, nil }

func NewForConfigOrDie(*rest.Config) *Client { return &Client{} }
