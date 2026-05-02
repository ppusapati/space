package services_test

import (
	"context"
	"strings"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/sat-fsw/internal/models"
	"github.com/ppusapati/space/services/sat-fsw/internal/services"
)

func TestRegisterFirmwareBuildRejectsEmptyTenant(t *testing.T) {
	s := services.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), services.RegisterFirmwareBuildInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestRegisterFirmwareBuildRejectsBadGitSHA(t *testing.T) {
	s := services.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), services.RegisterFirmwareBuildInput{
		TenantID:          ulid.New(),
		TargetPlatform:    "stm32-h7-cdh",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            "not-hex",
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 1024,
		ArtefactSHA256:    strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for bad git_sha")
	}
}

func TestRegisterFirmwareBuildRejectsBadSHA256(t *testing.T) {
	s := services.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), services.RegisterFirmwareBuildInput{
		TenantID:          ulid.New(),
		TargetPlatform:    "stm32-h7-cdh",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            strings.Repeat("a", 40),
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 1024,
		ArtefactSHA256:    "z",
	})
	if err == nil {
		t.Fatal("expected error for bad artefact_sha256")
	}
}

func TestRegisterFirmwareBuildRejectsZeroSize(t *testing.T) {
	s := services.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), services.RegisterFirmwareBuildInput{
		TenantID:          ulid.New(),
		TargetPlatform:    "x",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            strings.Repeat("a", 40),
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 0,
		ArtefactSHA256:    strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for zero artefact_size_bytes")
	}
}

func TestUpdateFirmwareBuildStatusRejectsUnspecified(t *testing.T) {
	s := services.New(nil)
	_, err := s.UpdateFirmwareBuildStatus(context.Background(), ulid.New(), models.BuildUnspecified, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestCreateDeploymentManifestRejectsEmptyAssignments(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateDeploymentManifest(context.Background(), services.CreateDeploymentManifestInput{
		TenantID:        ulid.New(),
		SatelliteID:     ulid.New(),
		ManifestVersion: "v1",
		Assignments:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error for empty assignments")
	}
}

func TestCreateDeploymentManifestRejectsBadULID(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateDeploymentManifest(context.Background(), services.CreateDeploymentManifestInput{
		TenantID:        ulid.New(),
		SatelliteID:     ulid.New(),
		ManifestVersion: "v1",
		Assignments:     map[string]string{"cdh": "not-a-ulid"},
	})
	if err == nil {
		t.Fatal("expected error for malformed firmware_build_id")
	}
}

func TestUpdateDeploymentManifestStatusRejectsUnspecified(t *testing.T) {
	s := services.New(nil)
	_, err := s.UpdateDeploymentManifestStatus(context.Background(), ulid.New(), models.DeploymentUnspecified, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListFirmwareBuildsForTenant(context.Background(), services.ListFirmwareBuildsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
	if _, _, err := s.ListDeploymentManifestsForTenant(context.Background(), services.ListDeploymentManifestsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
