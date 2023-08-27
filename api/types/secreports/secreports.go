/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secreports

import (
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/types/header"
	"github.com/gravitational/teleport/api/types/header/convert/legacy"
)

// SecurityReport is ...
type SecurityReport struct {
	// ResourceHeader is
	header.ResourceHeader
	// Spec is ...
	Spec SecurityReportSpec `json:"spec" yaml:"spec"`
}

// SecurityReportSpec is ...
type SecurityReportSpec struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Query is ..
	Desc string `json:"desc,omitempty" yaml:"desc,omitempty"`
	//Result is ...
	Queries []string `json:"queries,omitempty" yaml:"queries,omitempty"`
}

// AuditQuery is ...
type AuditQuery struct {
	// ResourceHeader is
	header.ResourceHeader
	// Spec is ...
	Spec AuditQuerySpec `json:"spec" yaml:"spec"`
}

// AuditQuerySpec is ...
type AuditQuerySpec struct {
	// Query is ..
	Query string `json:"query,omitempty" yaml:"query,omitempty"`
	// Desc is a audit query short description.
	Desc string `json:"desc,omitempty" yaml:"desc,omitempty"`
	// ExecutionIID audit query execution id.
	ExecutionID string `json:"execution_id,omitempty" yaml:"result,omitempty"`
}

// CheckAndSetDefaults validates fields and populates empty fields with default values.
func (a *AuditQuery) CheckAndSetDefaults() error {
	a.SetKind(types.KindAuditQuery)
	a.SetVersion(types.V1)

	if err := a.ResourceHeader.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func NewAuditQuery(metadata header.Metadata, spec AuditQuerySpec) (*AuditQuery, error) {
	secReport := &AuditQuery{
		ResourceHeader: header.ResourceHeaderFromMetadata(metadata),
		Spec:           spec,
	}
	if err := secReport.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return secReport, nil
}

func NewSecurityReport(metadata header.Metadata, spec SecurityReportSpec) (*SecurityReport, error) {
	secReport := &SecurityReport{
		ResourceHeader: header.ResourceHeaderFromMetadata(metadata),
		Spec:           spec,
	}
	if err := secReport.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return secReport, nil
}

// CheckAndSetDefaults validates fields and populates empty fields with default values.
func (a *SecurityReport) CheckAndSetDefaults() error {
	a.SetKind(types.KindSecurityReport)
	a.SetVersion(types.V1)

	if err := a.ResourceHeader.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

//func (a *SecurityReport) SetUpdateTime(days int, updatedAt time.Time) {
//	if a.Spec.UpdateTimeMap == nil {
//		a.Spec.UpdateTimeMap = make(map[int]string)
//	}
//	a.Spec.UpdateTimeMap[days] = updatedAt.UTC().Format(time.RFC3339)
//}

// GetMetadata returns metadata. This is specifically for conforming to the Resource interface,
// and should be removed when possible.
func (a *SecurityReport) GetMetadata() types.Metadata {
	return legacy.FromHeaderMetadata(a.Metadata)
}

// GetMetadata returns metadata. This is specifically for conforming to the Resource interface,
// and should be removed when possible.
func (a *AuditQuery) GetMetadata() types.Metadata {
	return legacy.FromHeaderMetadata(a.Metadata)
}

type SecurityReportStatus string

const (
	Unknown SecurityReportStatus = "UNKNOWN"
	Running SecurityReportStatus = "RUNNING"
	Failed  SecurityReportStatus = "FAILED"
	Ready   SecurityReportStatus = "READY"
)

type SecurityReportState struct {
	// ResourceHeader is
	header.ResourceHeader
	Spec SecurityReportStateSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type SecurityReportStateSpec struct {
	Status    SecurityReportStatus `json:"status,omitempty" yaml:"status,omitempty"`
	UpdatedAt string               `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// GetMetadata returns metadata. This is specifically for conforming to the Resource interface,
// and should be removed when possible.
func (a *SecurityReportState) GetMetadata() types.Metadata {
	return legacy.FromHeaderMetadata(a.Metadata)
}

// CheckAndSetDefaults validates fields and populates empty fields with default values.
func (a *SecurityReportState) CheckAndSetDefaults() error {
	a.SetKind(types.KindSecurityReportState)
	a.SetVersion(types.V1)

	if err := a.ResourceHeader.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func NewSecurityReportState(metadata header.Metadata, spec SecurityReportStateSpec) (*SecurityReportState, error) {
	secReport := &SecurityReportState{
		ResourceHeader: header.ResourceHeaderFromMetadata(metadata),
		Spec:           spec,
	}
	if err := secReport.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return secReport, nil
}
