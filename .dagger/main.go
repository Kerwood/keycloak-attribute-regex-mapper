package main

import (
	"context"
	"dagger/protocol-mapper/internal/dagger"
	"encoding/xml"
	"fmt"
)

type ProtocolMapper struct{}

func (m *ProtocolMapper) BuildJar(
	ctx context.Context,
	// +defaultPath="/"
	srcDirectory *dagger.Directory,
) *dagger.File {
	container := dag.Container().
		From("maven:3.8").
		WithMountedDirectory("/workdir", srcDirectory).
		WithWorkdir("/workdir").
		WithMountedCache("/root/maven/.m2", dag.CacheVolume("buildCache")).
		WithExec([]string{"mvn", "clean", "package"})

	daggerFile := container.File("/workdir/target/user-attribute-filter.jar")

	return daggerFile
}

// ############################################################ //
//                             Release                          //
// ############################################################ //

// Checks for version conflicts, builds jar file, creates Github release and uploads binary as asset.
func (m *ProtocolMapper) Release(
	ctx context.Context,
	// +defaultPath="/"
	srcDirectory *dagger.Directory,
	ghUser string,
	ghAuthToken *dagger.Secret,
) error {

	pomVersion, err := m.CheckVersionConflict(ctx, srcDirectory, ghAuthToken)
	if err != nil {
		return err
	}

	jar, err := m.BuildJar(ctx, srcDirectory).Export(ctx, "/user-attribute-filter.jar")
	if err != nil {
		return err
	}

	github, err := NewGithubClient().WithAuthToken(ctx, ghAuthToken)
	if err != nil {
		return err
	}

	release, err := github.CreateRelease(ctx, "v"+pomVersion)
	if err != nil {
		return err
	}

	err = github.UploadAsset(ctx, *release.ID, jar)
	if err != nil {
		return err
	}

	return nil
}

// ############################################################ //
//                   Check Version Conflict                     //
// ############################################################ //

// Gets the package version from Cargo.toml and validates the tag does not exist in the repo. Returns the Cargo version as string.
func (m *ProtocolMapper) CheckVersionConflict(
	ctx context.Context,
	// +defaultPath="/"
	srcDirectory *dagger.Directory,
	ghAuthToken *dagger.Secret,
) (string, error) {

	pomFile := srcDirectory.File("./pom.xml")
	pomVersion, err := m.getPomVersion(ctx, pomFile)
	if err != nil {
		return "", err
	}

	github, err := NewGithubClient().WithAuthToken(ctx, ghAuthToken)
	if err != nil {
		return "", err
	}

	tags, err := github.ListTags(ctx)
	if err != nil {
		return "", err
	}

	for _, x := range tags {
		tag := "v" + pomVersion
		if x == tag {
			return "", fmt.Errorf("Conflict: Git tag '%s' already exists and matches the declared version in Cargo.toml (%s)", tag, pomVersion)
		}
	}

	return pomVersion, nil
}

// ############################################################ //
//                   Get version from pom.xml                   //
// ############################################################ //

type Project struct {
	Version string `xml:"version"`
}

// Get the version from the pom.xml file
func (m *ProtocolMapper) getPomVersion(ctx context.Context, pomFile *dagger.File) (string, error) {
	pomContent, err := pomFile.Contents(ctx)
	if err != nil {
		return "", err
	}

	var project Project
	if err := xml.Unmarshal([]byte(pomContent), &project); err != nil {
		return "", nil
	}
	return project.Version, nil
}
