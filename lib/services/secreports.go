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

package services

import (
	"context"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types/secreports"
	"github.com/gravitational/teleport/lib/utils"
)

// SecReports is
type SecReports interface {
	GetAuditQuery(ctx context.Context, name string) (*secreports.AuditQuery, error)
	CreateAuditQuery(ctx context.Context, in *secreports.AuditQuery) error
	UpsertAuditQuery(ctx context.Context, in *secreports.AuditQuery) error
	ListAuditQuery(context.Context, int, string) ([]*secreports.AuditQuery, string, error)
	DeleteAuditQuery(ctx context.Context, name string) error
	CreateSecurityReports(ctx context.Context, item *secreports.SecurityReport) error
	UpsertSecurityReports(ctx context.Context, item *secreports.SecurityReport) error
	GetSecurityReport(ctx context.Context, name string) (*secreports.SecurityReport, error)
	ListSecurityReport(ctx context.Context, i int, token string) ([]*secreports.SecurityReport, string, error)
}

func MarshalAuditQuery(in *secreports.AuditQuery, opts ...MarshalOption) ([]byte, error) {
	if err := in.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if !cfg.PreserveResourceID {
		copy := *in
		copy.SetResourceID(0)
		in = &copy
	}
	return utils.FastMarshal(in)
}

func UnmarshalAuditQuery(data []byte, opts ...MarshalOption) (*secreports.AuditQuery, error) {
	if len(data) == 0 {
		return nil, trace.BadParameter("missing access list data")
	}
	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var out *secreports.AuditQuery
	if err := utils.FastUnmarshal(data, &out); err != nil {
		return nil, trace.BadParameter(err.Error())
	}
	if err := out.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	if cfg.ID != 0 {
		out.SetResourceID(cfg.ID)
	}
	if !cfg.Expires.IsZero() {
		out.SetExpiry(cfg.Expires)
	}
	return out, nil
}

func MarshalSecurityReport(in *secreports.SecurityReport, opts ...MarshalOption) ([]byte, error) {
	if err := in.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if !cfg.PreserveResourceID {
		copy := *in
		copy.SetResourceID(0)
		in = &copy
	}
	return utils.FastMarshal(in)
}

func UnmarshalSecurityReport(data []byte, opts ...MarshalOption) (*secreports.SecurityReport, error) {
	if len(data) == 0 {
		return nil, trace.BadParameter("missing access list data")
	}
	cfg, err := CollectOptions(opts)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var out *secreports.SecurityReport
	if err := utils.FastUnmarshal(data, &out); err != nil {
		return nil, trace.BadParameter(err.Error())
	}
	if err := out.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	if cfg.ID != 0 {
		out.SetResourceID(cfg.ID)
	}
	if !cfg.Expires.IsZero() {
		out.SetExpiry(cfg.Expires)
	}
	return out, nil
}
