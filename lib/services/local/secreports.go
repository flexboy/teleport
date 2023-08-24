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

package local

import (
	"context"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"

	secreportsv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/secreports/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/types/secreports"
	"github.com/gravitational/teleport/lib/backend"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local/generic"
)

const (
	AuditQueryPrefix     = "audit_query"
	SecurityReportPrefix = "security_report"
)

// SecReportsService ...
type SecReportsService struct {
	log               logrus.FieldLogger
	clock             clockwork.Clock
	auditQuerySvc     *generic.Service[*secreports.AuditQuery]
	securityReportSvc *generic.Service[*secreports.SecurityReport]

	secreportsv1.UnimplementedSecReportsServiceServer
}

// NewSecReportsService ...
func NewSecReportsService(backend backend.Backend, clock clockwork.Clock) (*SecReportsService, error) {
	auditQuerySvc, err := generic.NewService(&generic.ServiceConfig[*secreports.AuditQuery]{
		Backend:       backend,
		ResourceKind:  types.KindAuditQuery,
		BackendPrefix: AuditQueryPrefix,
		MarshalFunc:   services.MarshalAuditQuery,
		UnmarshalFunc: services.UnmarshalAuditQuery,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	securityReportSvc, err := generic.NewService(&generic.ServiceConfig[*secreports.SecurityReport]{
		Backend:       backend,
		ResourceKind:  types.KindSecurityReport,
		BackendPrefix: SecurityReportPrefix,
		MarshalFunc:   services.MarshalSecurityReport,
		UnmarshalFunc: services.UnmarshalSecurityReport,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &SecReportsService{
		log:               logrus.WithFields(logrus.Fields{trace.Component: "secreports:local-service"}),
		clock:             clock,
		auditQuerySvc:     auditQuerySvc,
		securityReportSvc: securityReportSvc,
	}, nil
}

// UpsertAuditQuery is ..
func (s *SecReportsService) UpsertAuditQuery(ctx context.Context, in *secreports.AuditQuery) error {
	if err := s.auditQuerySvc.UpsertResource(ctx, in); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (s *SecReportsService) CreateAuditQuery(ctx context.Context, in *secreports.AuditQuery) error {
	if err := s.auditQuerySvc.UpsertResource(ctx, in); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// GetAuditQuery is ..
func (s *SecReportsService) GetAuditQuery(ctx context.Context, name string) (*secreports.AuditQuery, error) {
	r, err := s.auditQuerySvc.GetResource(ctx, name)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return r, nil
}

func (s *SecReportsService) ListAuditQuery(ctx context.Context, pageSize int, nextToken string) ([]*secreports.AuditQuery, string, error) {
	items, nextToken, err := s.auditQuerySvc.ListResources(ctx, pageSize, nextToken)
	if err != nil {
		return nil, "", trace.Wrap(err)
	}
	return items, nextToken, err
}

func (s *SecReportsService) DeleteAuditQuery(ctx context.Context, name string) error {
	return trace.Wrap(s.auditQuerySvc.DeleteResource(ctx, name))
}

func (s *SecReportsService) CreateSecurityReports(ctx context.Context, item *secreports.SecurityReport) error {
	if err := s.securityReportSvc.CreateResource(ctx, item); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (s *SecReportsService) UpsertSecurityReports(ctx context.Context, item *secreports.SecurityReport) error {
	if err := s.securityReportSvc.UpsertResource(ctx, item); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (s *SecReportsService) GetSecurityReport(ctx context.Context, name string) (*secreports.SecurityReport, error) {
	r, err := s.securityReportSvc.GetResource(ctx, name)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return r, nil
}

func (s *SecReportsService) ListSecurityReport(ctx context.Context, i int, token string) ([]*secreports.SecurityReport, string, error) {
	items, nextToken, err := s.securityReportSvc.ListResources(ctx, i, token)
	if err != nil {
		return nil, "", trace.Wrap(err)
	}
	return items, nextToken, err
}
