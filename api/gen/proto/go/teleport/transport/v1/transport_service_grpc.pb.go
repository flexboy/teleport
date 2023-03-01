// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: teleport/transport/v1/transport_service.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TransportServiceClient is the client API for TransportService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TransportServiceClient interface {
	// GetClusterDetails provides cluster information that may affect how transport
	// should occur.
	GetClusterDetails(ctx context.Context, in *GetClusterDetailsRequest, opts ...grpc.CallOption) (*GetClusterDetailsResponse, error)
	// ProxySSH establishes an SSH connection to the target host over a bidirectional stream.
	//
	// The client must first send a DialTarget before the connection is established. Agent frames
	// will be populated if SSH Agent forwarding is enabled for the connection. SSH frames contain
	// raw SSH payload to be processed by an x/crypto/ssh.Client or x/crypto/ssh.Server.
	ProxySSH(ctx context.Context, opts ...grpc.CallOption) (TransportService_ProxySSHClient, error)
	// ProxyCluster establishes a connection to the target cluster.
	//
	// The client must first send a ProxyClusterRequest with the desired cluster name before the
	// connection is established. After which the connection can be used to construct a new
	// auth.Client to the tunneled cluster.
	ProxyCluster(ctx context.Context, opts ...grpc.CallOption) (TransportService_ProxyClusterClient, error)
}

type transportServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTransportServiceClient(cc grpc.ClientConnInterface) TransportServiceClient {
	return &transportServiceClient{cc}
}

func (c *transportServiceClient) GetClusterDetails(ctx context.Context, in *GetClusterDetailsRequest, opts ...grpc.CallOption) (*GetClusterDetailsResponse, error) {
	out := new(GetClusterDetailsResponse)
	err := c.cc.Invoke(ctx, "/teleport.transport.v1.TransportService/GetClusterDetails", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *transportServiceClient) ProxySSH(ctx context.Context, opts ...grpc.CallOption) (TransportService_ProxySSHClient, error) {
	stream, err := c.cc.NewStream(ctx, &TransportService_ServiceDesc.Streams[0], "/teleport.transport.v1.TransportService/ProxySSH", opts...)
	if err != nil {
		return nil, err
	}
	x := &transportServiceProxySSHClient{stream}
	return x, nil
}

type TransportService_ProxySSHClient interface {
	Send(*ProxySSHRequest) error
	Recv() (*ProxySSHResponse, error)
	grpc.ClientStream
}

type transportServiceProxySSHClient struct {
	grpc.ClientStream
}

func (x *transportServiceProxySSHClient) Send(m *ProxySSHRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *transportServiceProxySSHClient) Recv() (*ProxySSHResponse, error) {
	m := new(ProxySSHResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *transportServiceClient) ProxyCluster(ctx context.Context, opts ...grpc.CallOption) (TransportService_ProxyClusterClient, error) {
	stream, err := c.cc.NewStream(ctx, &TransportService_ServiceDesc.Streams[1], "/teleport.transport.v1.TransportService/ProxyCluster", opts...)
	if err != nil {
		return nil, err
	}
	x := &transportServiceProxyClusterClient{stream}
	return x, nil
}

type TransportService_ProxyClusterClient interface {
	Send(*ProxyClusterRequest) error
	Recv() (*ProxyClusterResponse, error)
	grpc.ClientStream
}

type transportServiceProxyClusterClient struct {
	grpc.ClientStream
}

func (x *transportServiceProxyClusterClient) Send(m *ProxyClusterRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *transportServiceProxyClusterClient) Recv() (*ProxyClusterResponse, error) {
	m := new(ProxyClusterResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TransportServiceServer is the server API for TransportService service.
// All implementations must embed UnimplementedTransportServiceServer
// for forward compatibility
type TransportServiceServer interface {
	// GetClusterDetails provides cluster information that may affect how transport
	// should occur.
	GetClusterDetails(context.Context, *GetClusterDetailsRequest) (*GetClusterDetailsResponse, error)
	// ProxySSH establishes an SSH connection to the target host over a bidirectional stream.
	//
	// The client must first send a DialTarget before the connection is established. Agent frames
	// will be populated if SSH Agent forwarding is enabled for the connection. SSH frames contain
	// raw SSH payload to be processed by an x/crypto/ssh.Client or x/crypto/ssh.Server.
	ProxySSH(TransportService_ProxySSHServer) error
	// ProxyCluster establishes a connection to the target cluster.
	//
	// The client must first send a ProxyClusterRequest with the desired cluster name before the
	// connection is established. After which the connection can be used to construct a new
	// auth.Client to the tunneled cluster.
	ProxyCluster(TransportService_ProxyClusterServer) error
	mustEmbedUnimplementedTransportServiceServer()
}

// UnimplementedTransportServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTransportServiceServer struct {
}

func (UnimplementedTransportServiceServer) GetClusterDetails(context.Context, *GetClusterDetailsRequest) (*GetClusterDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClusterDetails not implemented")
}
func (UnimplementedTransportServiceServer) ProxySSH(TransportService_ProxySSHServer) error {
	return status.Errorf(codes.Unimplemented, "method ProxySSH not implemented")
}
func (UnimplementedTransportServiceServer) ProxyCluster(TransportService_ProxyClusterServer) error {
	return status.Errorf(codes.Unimplemented, "method ProxyCluster not implemented")
}
func (UnimplementedTransportServiceServer) mustEmbedUnimplementedTransportServiceServer() {}

// UnsafeTransportServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TransportServiceServer will
// result in compilation errors.
type UnsafeTransportServiceServer interface {
	mustEmbedUnimplementedTransportServiceServer()
}

func RegisterTransportServiceServer(s grpc.ServiceRegistrar, srv TransportServiceServer) {
	s.RegisterService(&TransportService_ServiceDesc, srv)
}

func _TransportService_GetClusterDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClusterDetailsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransportServiceServer).GetClusterDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.transport.v1.TransportService/GetClusterDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransportServiceServer).GetClusterDetails(ctx, req.(*GetClusterDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TransportService_ProxySSH_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TransportServiceServer).ProxySSH(&transportServiceProxySSHServer{stream})
}

type TransportService_ProxySSHServer interface {
	Send(*ProxySSHResponse) error
	Recv() (*ProxySSHRequest, error)
	grpc.ServerStream
}

type transportServiceProxySSHServer struct {
	grpc.ServerStream
}

func (x *transportServiceProxySSHServer) Send(m *ProxySSHResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *transportServiceProxySSHServer) Recv() (*ProxySSHRequest, error) {
	m := new(ProxySSHRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _TransportService_ProxyCluster_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TransportServiceServer).ProxyCluster(&transportServiceProxyClusterServer{stream})
}

type TransportService_ProxyClusterServer interface {
	Send(*ProxyClusterResponse) error
	Recv() (*ProxyClusterRequest, error)
	grpc.ServerStream
}

type transportServiceProxyClusterServer struct {
	grpc.ServerStream
}

func (x *transportServiceProxyClusterServer) Send(m *ProxyClusterResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *transportServiceProxyClusterServer) Recv() (*ProxyClusterRequest, error) {
	m := new(ProxyClusterRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TransportService_ServiceDesc is the grpc.ServiceDesc for TransportService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TransportService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "teleport.transport.v1.TransportService",
	HandlerType: (*TransportServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetClusterDetails",
			Handler:    _TransportService_GetClusterDetails_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ProxySSH",
			Handler:       _TransportService_ProxySSH_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ProxyCluster",
			Handler:       _TransportService_ProxyCluster_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "teleport/transport/v1/transport_service.proto",
}
