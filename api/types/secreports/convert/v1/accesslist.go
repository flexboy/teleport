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

package v1

import (
	"github.com/gravitational/trace"

	secreportsv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/secreports/v1"
	headerv1 "github.com/gravitational/teleport/api/types/header/convert/v1"
	"github.com/gravitational/teleport/api/types/secreports"
)

func FromProto(in *secreportsv1.AuditQuery) (*secreports.AuditQuery, error) {
	spec := secreports.AuditQuerySpec{
		Query: in.GetSpec().GetQuery(),
		Desc:  in.GetSpec().GetDesc(),
	}
	out, err := secreports.NewAuditQuery(headerv1.FromMetadataProto(in.Header.Metadata), spec)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return out, nil
}

func ToProto(in *secreports.AuditQuery) *secreportsv1.AuditQuery {
	return &secreportsv1.AuditQuery{
		Header: headerv1.ToResourceHeaderProto(in.ResourceHeader),
		Spec: &secreportsv1.AuditQuerySpec{
			Query: in.Spec.Query,
			Desc:  in.Spec.Desc,
		},
	}
}

func ReportFromProto(in *secreportsv1.SecurityReport) (*secreports.SecurityReport, error) {
	spec := secreports.SecurityReportSpec{
		Name:    in.GetSpec().GetName(),
		Desc:    in.GetSpec().GetDesc(),
		Queries: in.GetSpec().GetQueries(),
	}
	out, err := secreports.NewSecurityReport(headerv1.FromMetadataProto(in.Header.Metadata), spec)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return out, nil
}

func ReportToProto(in *secreports.SecurityReport) *secreportsv1.SecurityReport {
	return &secreportsv1.SecurityReport{
		Header: headerv1.ToResourceHeaderProto(in.ResourceHeader),
		Spec: &secreportsv1.SecurityReportSpec{
			Name:    in.GetName(),
			Desc:    in.Spec.Desc,
			Queries: in.Spec.Queries,
		},
	}
}
