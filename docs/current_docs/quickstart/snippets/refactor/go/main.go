package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

type HelloDagger struct{}

// Tests, builds, packages and publishes the application
func (m *HelloDagger) Ci(ctx context.Context, source *Directory) (string, error) {
	// run tests
	_, err := m.Test(ctx, source)
	if err != nil {
		return "", err
	}
	// create and publish a container with the build output
	address, err := m.Package(source).
		Publish(ctx, fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000))) //#nosec
	if err != nil {
		return "", err
	}
	return address, nil
}

// Returns a container with the production build
func (m *HelloDagger) Package(source *Directory) *Container {
	return dag.Container().From("nginx:1.25-alpine").
		WithDirectory("/usr/share/nginx/html", m.Build(source)).
		WithExposedPort(80)
}

// Returns a directory with the production build
func (m *HelloDagger) Build(source *Directory) *Directory {
	return dag.Node(NodeOpts{Version: "21"}).
		WithNpm().
		WithSource(source).
		Install(nil).
		Commands().
		Run([]string{"build"}).
		Directory("./dist")
}

// Returns the result of running unit tests
func (m *HelloDagger) Test(ctx context.Context, source *Directory) (string, error) {
	return dag.Node(NodeOpts{Version: "21"}).
		WithNpm().
		WithSource(source).
		Install(nil).
		Commands().
		Run([]string{"test:unit", "run"}).
		Stdout(ctx)
}