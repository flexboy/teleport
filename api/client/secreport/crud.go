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

package secreport

import (
	"context"
	"time"

	"github.com/gravitational/trace"
	"github.com/gravitational/trace/trail"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/gravitational/teleport/api/gen/proto/go/teleport/secreports/v1"
	"github.com/gravitational/teleport/api/types/secreports"
	v1 "github.com/gravitational/teleport/api/types/secreports/convert/v1"
)

func (c *Client) GetAuditQuery(ctx context.Context, name string) (*secreports.AuditQuery, error) {
	resp, err := c.grpcClient.GetAuditQuery(ctx, &pb.GetAuditQueryRequest{Name: name})
	if err != nil {
		return nil, trail.FromGRPC(err)
	}
	out, err := v1.FromProto(resp)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return out, nil
}

func (c *Client) UpsertAuditQuery(ctx context.Context, in *secreports.AuditQuery) error {
	_, err := c.grpcClient.UpsertAuditQuery(ctx, &pb.UpsertAuditQueryRequest{AuditQuery: v1.ToProto(in)})
	if err != nil {
		return trail.FromGRPC(err)
	}
	return nil
}

func (c *Client) ListAuditQuery(ctx context.Context) ([]*secreports.AuditQuery, error) {
	var resources []*pb.AuditQuery
	nextKey := ""
	for {
		resp, err := c.grpcClient.ListAuditQuery(ctx, &pb.ListAuditQueryRequest{
			PageSize:  0,
			PageToken: nextKey,
		})
		if err != nil {
			return nil, trace.Wrap(err)
		}
		resources = append(resources, resp.GetQueries()...)
		if nextKey == "" {
			break
		}
	}
	out := make([]*secreports.AuditQuery, 0, len(resources))
	for _, v := range resources {
		item, err := v1.FromProto(v)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		out = append(out, item)
	}
	return out, nil
}

func (c *Client) DeleteAuditQuery(ctx context.Context, name string) error {
	_, err := c.grpcClient.DeleteAuditQuery(ctx, &pb.DeleteAuditQueryRequest{Name: name})
	if err != nil {
		return trail.FromGRPC(err)
	}
	return nil
}

func (c *Client) UpsertSecurityReports(ctx context.Context, item *secreports.SecurityReport) error {
	_, err := c.grpcClient.UpsertSecurityReport(ctx, &pb.UpsertSecurityReportRequest{Report: v1.ReportToProto(item)})
	if err != nil {
		return trail.FromGRPC(err)
	}
	return nil
}

func (c *Client) GetSecurityReport(ctx context.Context, name string) (*secreports.SecurityReport, error) {
	resp, err := c.grpcClient.GetSecurityReport(ctx, &pb.GetSecurityReportRequest{Name: name})
	if err != nil {
		return nil, trail.FromGRPC(err)
	}

	out, err := v1.ReportFromProto(resp)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return out, nil
}

func (c *Client) ListSecurityReport(ctx context.Context) ([]*secreports.SecurityReport, error) {
	var resources []*pb.SecurityReport
	nextKey := ""
	for {
		resp, err := c.grpcClient.ListSecurityReport(ctx, &pb.ListSecurityReportRequest{
			PageSize:  0,
			PageToken: nextKey,
		})
		if err != nil {
			return nil, trace.Wrap(err)
		}
		resources = append(resources, resp.GetReports()...)
		if nextKey == "" {
			break
		}
	}

	out := make([]*secreports.SecurityReport, 0, len(resources))
	for _, v := range resources {
		item, err := v1.ReportFromProto(v)
		if err != nil {
			return nil, trace.Wrap(err)

		}
		out = append(out, item)
	}
	return out, nil
}

func (c *Client) GetSecurityReportsDetails(ctx context.Context, name string, from, to time.Time) (*pb.GetSecurityReportDetailsResponse, error) {
	resp, err := c.grpcClient.GetSecurityReportDetails(ctx, &pb.GetSecurityReportDetailsRequest{
		Name:      name,
		StartData: timestamppb.New(from),
		EndDate:   timestamppb.New(to),
	})
	if err != nil {
		return nil, trail.FromGRPC(err)
	}
	return resp, nil
}

func (c *Client) RunSecurityReport(ctx context.Context, name string) error {
	_, err := c.grpcClient.RunSecurityReport(ctx, &pb.RunSecurityReportRequest{Name: name})
	if err != nil {
		return trail.FromGRPC(err)
	}
	return nil
}
