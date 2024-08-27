package main

import (
	"context"
	"dagger/compatcheck/internal/dagger"
	"dagger/compatcheck/schemadiff"
	"fmt"
	"runtime"

	"github.com/dagger/dagger/dagql/introspection"
	"github.com/moby/buildkit/identity"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/rand"
	"golang.org/x/mod/semver"
)

type Compatcheck struct{}

// Compare the schema of given module with different versions of engine
func (m *Compatcheck) Run(ctx context.Context,
	// ref of the module to compare
	module,
	// version of engine to compare with
	versionA,
	// version of engine to compare to
	versionB string,
	// +optional
	// only required when one of the version to compare is 'dev'
	source *dagger.Directory,
) error {
	if (versionA == "dev" || versionB == "dev") && source == nil {
		return fmt.Errorf("--source flag is required when one of the requested engine version is 'dev'")
	}

	baseSchemaA, schemaA, err := m.getSchemaForModuleForEngineVersion(ctx, module, versionA, source)
	if err != nil {
		return err
	}

	baseSchemaB, schemaB, err := m.getSchemaForModuleForEngineVersion(ctx, module, versionB, source)
	if err != nil {
		return err
	}

	diff, err := schemadiff.Do(baseSchemaA, baseSchemaB, schemaA, schemaB)
	if err != nil {
		return err
	}

	if diff != "" {
		return fmt.Errorf("%s", diff)
	}

	return nil
}

// setup dagger engine/client with requested version and
// fetches schema using dagger query
func (m *Compatcheck) getSchemaForModuleForEngineVersion(ctx context.Context, module, engineVersion string, source *dagger.Directory) (string, string, error) {
	var engineSvc *dagger.Service
	var client *dagger.Container
	var err error

	if engineVersion == "dev" {
		client = devEngineAndClient(source)
	} else {
		engineSvc = engineServiceWithVersion(engineVersion)
		client, err = engineClientContainerWithVersion(ctx, engineSvc, engineVersion)
		if err != nil {
			return "", "", err
		}
	}

	baseIntrospection, err := client.WithNewFile("/base-schema-query.graphql", introspection.Query).
		WithExec([]string{"dagger", "query", "--doc", "/base-schema-query.graphql"}).
		Stdout(ctx)

	if err != nil {
		return "", "", err
	}

	withModuleIntrospection, err := client.WithNewFile("/schema-query.graphql", introspection.Query).
		WithExec([]string{"dagger", "query", "-m", module, "--doc", "/schema-query.graphql"}).
		Stdout(ctx)

	if err != nil {
		return "", "", err
	}

	return gjson.Get(baseIntrospection, "__schema").String(), gjson.Get(withModuleIntrospection, "__schema").String(), nil
}

// returns a container with requested version of dagger cli
func engineClientContainerWithVersion(ctx context.Context, devEngine *dagger.Service, version string) (*dagger.Container, error) {
	endpoint, err := devEngine.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 1234, Scheme: "tcp"})
	if err != nil {
		return nil, err
	}

	// since release v0.12.5, dagger cli is bundled with the docker image of engine
	if semver.Compare(version, "v0.12.5") >= 0 {
		return dag.Container().From(fmt.Sprintf("ghcr.io/dagger/engine:%s", version)).
			WithServiceBinding("dev-engine", devEngine).
			WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", endpoint), nil
	}

	releaseArtifactName := fmt.Sprintf("dagger_%s_%s_%s", version, runtime.GOOS, runtime.GOARCH)
	releaseArtifactTarFile := fmt.Sprintf("%s.tar.gz", releaseArtifactName)
	releaseArtifactDownloadLink := fmt.Sprintf("https://github.com/dagger/dagger/releases/download/%s/%s", version, releaseArtifactTarFile)

	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"wget", releaseArtifactDownloadLink, "-O", releaseArtifactTarFile}).
		WithExec([]string{"tar", "-xvf", releaseArtifactTarFile}).
		WithExec([]string{"mv", "dagger", "/usr/bin/dagger"}).
		WithServiceBinding("dev-engine", devEngine).
		WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", endpoint), nil
}

// returns a service with requested version of the dagger engine
func engineServiceWithVersion(version string, withs ...func(*dagger.Container) *dagger.Container) *dagger.Service {
	ctr := dag.Container().From(fmt.Sprintf("ghcr.io/dagger/engine:%s", version))
	for _, with := range withs {
		ctr = with(ctr)
	}

	deviceName, cidr := getUniqueNestedEngineNetwork()
	return ctr.
		WithMountedCache("/var/lib/dagger", dag.CacheVolume("dagger-dev-engine-state-"+identity.NewID())).
		WithExposedPort(1234, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithExec([]string{
			"--addr", "tcp://0.0.0.0:1234",
			"--addr", "unix:///var/run/buildkit/buildkitd.sock",
			// // avoid network conflicts with other tests
			"--network-name", deviceName,
			"--network-cidr", cidr,
		}, dagger.ContainerWithExecOpts{
			UseEntrypoint:            true,
			InsecureRootCapabilities: true,
		}).AsService()
}

// returns a container with dev version of dagger engine and cli
// requires path to root of dagger repository to be provided using
// --source flag
func devEngineAndClient(source *dagger.Directory) *dagger.Container {
	return dag.DaggerDev(source).Dev()
}

// creates a network CIDR to use for running the engine
func getUniqueNestedEngineNetwork() (deviceName string, cidr string) {
	random := rand.Intn(240)
	return fmt.Sprintf("dagger%d", random), fmt.Sprintf("10.89.%d.0/24", random)
}