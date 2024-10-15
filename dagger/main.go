// // A generated module for BifrostBackups functions
// //
// // This module has been generated via dagger init and serves as a reference to
// // basic module structure as you get started with Dagger.
// //
// // Two functions have been pre-created. You can modify, delete, or add to them,
// // as needed. They demonstrate usage of arguments and return types using simple
// // echo and grep commands. The functions can be called from the dagger CLI or
// // from one of the SDKs.
// //
// // The first line in this comment block is a short description line and the
// // rest is a long description with more detail on the module's purpose or usage,
// // if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/bifrost-backups/internal/dagger"
	"os"
)

type BifrostBackups struct{}

// Return the result of running unit tests
func (m *BifrostBackups) Test(ctx context.Context, source *dagger.Directory) (string, error) {
	return m.BuildEnv(source).
		WithExec([]string{"go", "install", "github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest"}).
		WithExec([]string{"sh", "-c", "go test -json -v ./... 2>&1 | tee /tmp/gotest.log | gotestfmt"}).
		Stdout(ctx)
}

// Build a ready-to-use development environment
func (m *BifrostBackups) BuildEnv(source *dagger.Directory) *dagger.Container {
	goCache := dag.CacheVolume("go")
	return dag.Container().
		From("golang:1.23-alpine"). // Updated to Go 1.23
		// From("ghcr.io/goreleaser/goreleaser-cross:v1.23"). // Updated to Go 1.23
		WithMountedDirectory("/app", source).
		WithWorkdir("/app").
		WithMountedCache("/go/pkg/mod", goCache).
		// WithExec([]string{"apt-get", "update"}).
		// WithExec([]string{"apt-get", "install", "-y", "gcc", "gcc-multilib"}).            // Added gcc and libc-dev
		WithExec([]string{"apk", "add", "--no-cache", "git", "make", "gcc", "musl-dev"}). // Alpine package names
		WithEnvVariable("CGO_ENABLED", "1").                                              // Enable CGO
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "build", "-o", "bin/bifrost-backups", "."})
}

// func (m *BifrostBackups) Release(ctx context.Context, source *dagger.Directory) (string, error) {
// 	return m.BuildEnv(source).
// 		WithMountedFile("/app/.goreleaser.yml", client.Ho.File(".goreleaser.yml")).
// 		WithEnvVariable("GITHUB_TOKEN", dag.SetSecret("GITHUB_TOKEN").Secret()).
// 		WithExec([]string{"goreleaser", "release", "--clean"}).
// 		Stdout(ctx)
// }

func (m *BifrostBackups) Release(ctx context.Context, source *dagger.Directory) (string, error) {
	// Run gorelaser release to check if the binary compiles in all platforms
	goreleaser := m.goreleaserContainer(ctx, source).
		WithExec([]string{"goreleaser", "release", "--snapshot", "--clean"})

	// Return any errors from the goreleaser build
	_, err := goreleaser.Stderr(ctx)

	if err != nil {
		return "", err
	}

	return "Release tasks completed successfully!", nil
}

func (m *BifrostBackups) goreleaserContainer(ctx context.Context, source *dagger.Directory) *dagger.Container {
	// Run go build to check if the binary compiles
	return m.BuildEnv(source).
		WithExec([]string{"go", "install", "github.com/goreleaser/goreleaser/v2@latest"}).
		WithMountedFile("/app/.goreleaser.yml", source.File(".github/.goreleaser.yaml")).
		WithEnvVariable("GITHUB_TOKEN", os.Getenv("GITHUB_TOKEN"))
}
