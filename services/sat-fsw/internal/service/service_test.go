package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-fsw/internal/service"
)

func TestRegisterFirmwareBuildRejectsEmptyTenant(t *testing.T) {
	s := service.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), service.RegisterFirmwareBuildInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRegisterFirmwareBuildRejectsBadGitSHA(t *testing.T) {
	s := service.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), service.RegisterFirmwareBuildInput{
		TenantID:          uuid.New(),
		TargetPlatform:    "stm32-h7-cdh",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            "not-hex",
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 1024,
		ArtefactSHA256:    strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for bad git sha")
	}
}

func TestRegisterFirmwareBuildRejectsBadArtefactSHA(t *testing.T) {
	s := service.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), service.RegisterFirmwareBuildInput{
		TenantID:          uuid.New(),
		TargetPlatform:    "stm32-h7-cdh",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            strings.Repeat("a", 40),
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 1024,
		ArtefactSHA256:    strings.Repeat("z", 64),
	})
	if err == nil {
		t.Fatal("expected error for non-hex artefact sha")
	}
}

func TestRegisterFirmwareBuildRejectsZeroSize(t *testing.T) {
	s := service.New(nil)
	_, err := s.RegisterFirmwareBuild(context.Background(), service.RegisterFirmwareBuildInput{
		TenantID:          uuid.New(),
		TargetPlatform:    "stm32-h7-cdh",
		Subsystem:         "cdh",
		Version:           "1.0.0",
		GitSHA:            strings.Repeat("a", 40),
		ArtefactURI:       "s3://x/y.bin",
		ArtefactSizeBytes: 0,
		ArtefactSHA256:    strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for zero size")
	}
}

func TestUpdateFirmwareBuildStatusRejectsUnspecified(t *testing.T) {
	s := service.New(nil)
	_, err := s.UpdateFirmwareBuildStatus(context.Background(), uuid.New(), 0, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestCreateDeploymentManifestRejectsEmptyAssignments(t *testing.T) {
	s := service.New(nil)
	_, err := s.CreateDeploymentManifest(context.Background(), service.CreateDeploymentManifestInput{
		TenantID:        uuid.New(),
		SatelliteID:     uuid.New(),
		ManifestVersion: "v1",
		Assignments:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected error for empty assignments")
	}
}

func TestCreateDeploymentManifestRejectsBadUUID(t *testing.T) {
	s := service.New(nil)
	_, err := s.CreateDeploymentManifest(context.Background(), service.CreateDeploymentManifestInput{
		TenantID:        uuid.New(),
		SatelliteID:     uuid.New(),
		ManifestVersion: "v1",
		Assignments:     map[string]string{"cdh": "not-a-uuid"},
	})
	if err == nil {
		t.Fatal("expected error for malformed firmware_build_id")
	}
}

func TestUpdateDeploymentManifestStatusRejectsUnspecified(t *testing.T) {
	s := service.New(nil)
	_, err := s.UpdateDeploymentManifestStatus(context.Background(), uuid.New(), 0, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := service.New(nil)
	if _, err := s.ListFirmwareBuilds(context.Background(), service.ListFirmwareBuildsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
	if _, err := s.ListDeploymentManifests(context.Background(), service.ListDeploymentManifestsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
