package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gs-mc/internal/models"
	"github.com/ppusapati/space/services/gs-mc/internal/services"
)

func TestCreateGroundStationRejectsEmptyTenant(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateGroundStation(context.Background(), services.CreateGroundStationInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateGroundStationRejectsBadCountryCode(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateGroundStation(context.Background(), services.CreateGroundStationInput{
		TenantID:    ulid.New(),
		Slug:        "svalbard-1",
		Name:        "Svalbard",
		CountryCode: "NORWAY",
	})
	if err == nil {
		t.Fatal("expected error for bad country_code")
	}
}

func TestCreateGroundStationRejectsBadLatitude(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateGroundStation(context.Background(), services.CreateGroundStationInput{
		TenantID:    ulid.New(),
		Slug:        "x",
		Name:        "x",
		CountryCode: "NO",
		LatitudeDeg: 91,
	})
	if err == nil {
		t.Fatal("expected error for latitude > 90")
	}
}

func TestCreateAntennaRejectsEmptyIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateAntenna(context.Background(), services.CreateAntennaInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestCreateAntennaRejectsUnspecifiedBand(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateAntenna(context.Background(), services.CreateAntennaInput{
		TenantID:  ulid.New(),
		StationID: ulid.New(),
		Slug:      "ant-1",
		Name:      "Antenna 1",
	})
	if err == nil {
		t.Fatal("expected error for unspecified band")
	}
}

func TestCreateAntennaRejectsInvertedFreqRange(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateAntenna(context.Background(), services.CreateAntennaInput{
		TenantID:     ulid.New(),
		StationID:    ulid.New(),
		Slug:         "ant-1",
		Name:         "Antenna 1",
		Band:         models.BandX,
		Polarization: models.PolRHCP,
		MinFreqHz:    9_000_000_000,
		MaxFreqHz:    8_000_000_000,
	})
	if err == nil {
		t.Fatal("expected error for max < min freq")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListGroundStationsForTenant(context.Background(), ulid.Zero, 0, 10); err == nil {
		t.Fatal("expected error for nil tenant on stations")
	}
	if _, _, err := s.ListAntennasForTenant(context.Background(), services.ListAntennasInput{}); err == nil {
		t.Fatal("expected error for nil tenant on antennas")
	}
}
