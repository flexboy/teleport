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

	"github.com/gravitational/trace/trail"

	pb "github.com/gravitational/teleport/api/gen/proto/go/teleport/secreports/v1"
)

type Client struct {
	grpcClient pb.SecReportsServiceClient
}

// NewClient creates a new Okta client.
func NewClient(grpcClient pb.SecReportsServiceClient) *Client {
	return &Client{
		grpcClient: grpcClient,
	}
}

func (c *Client) GetSchema(ctx context.Context) (*pb.GetSchemaResponse, error) {
	resp, err := c.grpcClient.GetSchema(ctx, &pb.GetSchemaRequest{})
	if err != nil {
		return nil, trail.FromGRPC(err)
	}
	return resp, nil
}

func (c *Client) RunAuditQuery(ctx context.Context, queryText string, queryName string, days int) (*pb.RunAuditQueryResponse, error) {
	var req pb.RunAuditQueryRequest
	if queryText != "" {
		req.Query = &pb.RunAuditQueryRequest_SqlText{SqlText: queryText}
	} else {
		req.Query = &pb.RunAuditQueryRequest_Name{Name: queryName}
	}
	req.Days = int32(days)

	resp, err := c.grpcClient.RunAuditQuery(ctx, &req)
	if err != nil {
		return nil, trail.FromGRPC(err)
	}
	return resp, nil
}
