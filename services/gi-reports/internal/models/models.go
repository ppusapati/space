// Package models holds gi-reports domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// ReportFormat mirrors gireportsv1.ReportFormat.
type ReportFormat int32

const (
	FormatUnspecified ReportFormat = 0
	FormatPDF         ReportFormat = 1
	FormatHTML        ReportFormat = 2
	FormatDOCX        ReportFormat = 3
	FormatXLSX        ReportFormat = 4
)

// ReportStatus mirrors gireportsv1.ReportStatus.
type ReportStatus int32

const (
	StatusUnspecified ReportStatus = 0
	StatusQueued      ReportStatus = 1
	StatusGenerating  ReportStatus = 2
	StatusCompleted   ReportStatus = 3
	StatusFailed      ReportStatus = 4
	StatusCanceled    ReportStatus = 5
)

// ReportTemplate is one report-template record.
type ReportTemplate struct {
	ID               ulid.ID
	TenantID         ulid.ID
	Slug             string
	Name             string
	Description      string
	TemplateURI      string
	Format           ReportFormat
	ParametersSchema string
	Active           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        string
	UpdatedBy        string
}

// Report is one generated report.
type Report struct {
	ID             ulid.ID
	TenantID       ulid.ID
	TemplateID     ulid.ID
	Status         ReportStatus
	ParametersJSON string
	OutputURI      string
	ErrorMessage   string
	StartedAt      time.Time
	FinishedAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	UpdatedBy      string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
