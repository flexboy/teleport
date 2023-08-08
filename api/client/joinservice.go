/*
Copyright 2022 Gravitational, Inc.

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

package client

import (
	"context"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
)

// JoinServiceClient is a client for the JoinService, which runs on both the
// auth and proxy.
type JoinServiceClient struct {
	grpcClient proto.JoinServiceClient
}

// NewJoinServiceClient returns a new JoinServiceClient wrapping the given grpc
// client.
func NewJoinServiceClient(grpcClient proto.JoinServiceClient) *JoinServiceClient {
	return &JoinServiceClient{
		grpcClient: grpcClient,
	}
}

// RegisterIAMChallengeResponseFunc is a function type meant to be passed to
// RegisterUsingIAMMethod. It must return a *proto.RegisterUsingIAMMethodRequest
// for a given challenge, or an error.
type RegisterIAMChallengeResponseFunc func(challenge string) (*proto.RegisterUsingIAMMethodRequest, error)

// RegisterAzureChallengeResponseFunc is a function type meant to be passed to
// RegisterUsingAzureMethod. It must return a
// *proto.RegisterUsingAzureMethodRequest for a given challenge, or an error.
type RegisterAzureChallengeResponseFunc func(challenge string) (*proto.RegisterUsingAzureMethodRequest, error)

// RegisterUsingIAMMethod registers the caller using the IAM join method and
// returns signed certs to join the cluster.
//
// The caller must provide a ChallengeResponseFunc which returns a
// *types.RegisterUsingTokenRequest with a signed sts:GetCallerIdentity request
// including the challenge as a signed header.
func (c *JoinServiceClient) RegisterUsingIAMMethod(ctx context.Context, challengeResponse RegisterIAMChallengeResponseFunc) (*proto.Certs, error) {
	// Make sure the gRPC stream is closed when this returns
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// initiate the streaming rpc
	iamJoinClient, err := c.grpcClient.RegisterUsingIAMMethod(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// wait for the challenge string from auth
	challenge, err := iamJoinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// get challenge response from the caller
	req, err := challengeResponse(challenge.Challenge)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// forward the challenge response from the caller to auth
	if err := iamJoinClient.Send(req); err != nil {
		return nil, trace.Wrap(err)
	}

	// wait for the certs from auth and return to the caller
	certsResp, err := iamJoinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return certsResp.Certs, nil
}

// RegisterUsingAzureMethod registers the caller using the Azure join method and
// returns signed certs to join the cluster.
//
// The caller must provide a ChallengeResponseFunc which returns a
// *proto.RegisterUsingAzureMethodRequest with a signed attested data document
// including the challenge as a nonce.
func (c *JoinServiceClient) RegisterUsingAzureMethod(ctx context.Context, challengeResponse RegisterAzureChallengeResponseFunc) (*proto.Certs, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	azureJoinClient, err := c.grpcClient.RegisterUsingAzureMethod(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	challenge, err := azureJoinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	req, err := challengeResponse(challenge.Challenge)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := azureJoinClient.Send(req); err != nil {
		return nil, trace.Wrap(err)
	}

	certsResp, err := azureJoinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return certsResp.Certs, nil
}

// KubernetesRemoteChallengeSolver
type KubernetesRemoteChallengeSolver func(audience string) (jwt string, err error)

// RegisterUsingKubernetesRemoteMethod
func (c *JoinServiceClient) RegisterUsingKubernetesRemoteMethod(ctx context.Context, tokenReq *types.RegisterUsingTokenRequest, solve KubernetesRemoteChallengeSolver) (*proto.Certs, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	joinClient, err := c.grpcClient.RegisterUsingKubernetesRemoteMethod(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// 1. Send initial request to server describing which token we wish to use.
	if err := joinClient.Send(&proto.RegisterUsingKubernetesRemoteMethodRequest{
		Payload: &proto.RegisterUsingKubernetesRemoteMethodRequest_RegisterUsingTokenRequest{
			RegisterUsingTokenRequest: tokenReq,
		},
	}); err != nil {
		return nil, trace.Wrap(err)
	}

	// 2. Receive the challenge audience from the server.
	resp, err := joinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	challenge := resp.GetChallengeAudience()
	if challenge == "" {
		return nil, trace.BadParameter("received empty challenge audience from server")
	}

	// 3. Solve the challenge using the provided callback func
	jwt, err := solve(challenge)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// 4. Send the solution to the server.
	if err := joinClient.Send(&proto.RegisterUsingKubernetesRemoteMethodRequest{
		Payload: &proto.RegisterUsingKubernetesRemoteMethodRequest_ChallengeSolutionJwt{
			ChallengeSolutionJwt: jwt,
		},
	}); err != nil {
		return nil, trace.Wrap(err)
	}

	// 5/
	resp, err = joinClient.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	certs := resp.GetCerts()
	if certs == nil {
		return nil, trace.BadParameter("did not receive expected certs from server")
	}

	return certs, nil
}
