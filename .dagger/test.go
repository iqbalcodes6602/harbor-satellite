package main

import (
	"context"

	"dagger/harbor-satellite/internal/dagger"
)

type Module struct{}

// =====================
// Unit Test Report
// =====================
func (m *Module) TestReport(
	ctx context.Context,
	source *dagger.Directory,
) (*dagger.File, error) {

	reportName := "TestReport.json"

	container := dag.Container().
		From("golang:1.22-alpine").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build")).
		WithEnvVariable("GOCACHE", "/go/build-cache").
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@latest"}).
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{
			"gotestsum",
			"--jsonfile",
			reportName,
			"./...",
		})

	return container.File(reportName), nil
}

// =====================
// Coverage Raw Output
// =====================
func (m *Module) TestCoverage(
	ctx context.Context,
	source *dagger.Directory,
) (*dagger.File, error) {

	coverage := "coverage.out"

	container := dag.Container().
		From("golang:1.22-alpine").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build")).
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{
			"go", "test", "./...",
			"-coverprofile=" + coverage,
		})

	return container.File(coverage), nil
}

// =====================
// Coverage Markdown Report
// =====================
func (m *Module) TestCoverageReport(
	ctx context.Context,
	source *dagger.Directory,
) (*dagger.File, error) {

	report := "coverage-report.md"
	coverage := "coverage.out"

	container := dag.Container().
		From("golang:1.22-alpine").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"apk", "add", "--no-cache", "bc"}).
		WithExec([]string{
			"go", "test", "./...",
			"-coverprofile=" + coverage,
		})

	return container.WithExec([]string{"sh", "-c", `
		echo "<h2> 📊 Test Coverage</h2>" > ` + report + `
		total=$(go tool cover -func=` + coverage + ` | grep total: | grep -Eo '[0-9]+\.[0-9]+')
		echo "<b>Total Coverage:</b> $total%" >> ` + report + `
		echo "<details><summary>Details</summary><pre>" >> ` + report + `
		go tool cover -func=` + coverage + ` >> ` + report + `
		echo "</pre></details>" >> ` + report + `
	`}).File(report), nil
}
