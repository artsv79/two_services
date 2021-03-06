// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cache.proto

package api

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type RandomDataArg struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RandomDataArg) Reset()         { *m = RandomDataArg{} }
func (m *RandomDataArg) String() string { return proto.CompactTextString(m) }
func (*RandomDataArg) ProtoMessage()    {}
func (*RandomDataArg) Descriptor() ([]byte, []int) {
	return fileDescriptor_5fca3b110c9bbf3a, []int{0}
}

func (m *RandomDataArg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RandomDataArg.Unmarshal(m, b)
}
func (m *RandomDataArg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RandomDataArg.Marshal(b, m, deterministic)
}
func (m *RandomDataArg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RandomDataArg.Merge(m, src)
}
func (m *RandomDataArg) XXX_Size() int {
	return xxx_messageInfo_RandomDataArg.Size(m)
}
func (m *RandomDataArg) XXX_DiscardUnknown() {
	xxx_messageInfo_RandomDataArg.DiscardUnknown(m)
}

var xxx_messageInfo_RandomDataArg proto.InternalMessageInfo

type RandomData struct {
	Str                  string   `protobuf:"bytes,1,opt,name=str,proto3" json:"str,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RandomData) Reset()         { *m = RandomData{} }
func (m *RandomData) String() string { return proto.CompactTextString(m) }
func (*RandomData) ProtoMessage()    {}
func (*RandomData) Descriptor() ([]byte, []int) {
	return fileDescriptor_5fca3b110c9bbf3a, []int{1}
}

func (m *RandomData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RandomData.Unmarshal(m, b)
}
func (m *RandomData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RandomData.Marshal(b, m, deterministic)
}
func (m *RandomData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RandomData.Merge(m, src)
}
func (m *RandomData) XXX_Size() int {
	return xxx_messageInfo_RandomData.Size(m)
}
func (m *RandomData) XXX_DiscardUnknown() {
	xxx_messageInfo_RandomData.DiscardUnknown(m)
}

var xxx_messageInfo_RandomData proto.InternalMessageInfo

func (m *RandomData) GetStr() string {
	if m != nil {
		return m.Str
	}
	return ""
}

func init() {
	proto.RegisterType((*RandomDataArg)(nil), "api.RandomDataArg")
	proto.RegisterType((*RandomData)(nil), "api.RandomData")
}

func init() { proto.RegisterFile("cache.proto", fileDescriptor_5fca3b110c9bbf3a) }

var fileDescriptor_5fca3b110c9bbf3a = []byte{
	// 125 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0x4e, 0x4c, 0xce,
	0x48, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2c, 0xc8, 0x54, 0xe2, 0xe7, 0xe2,
	0x0d, 0x4a, 0xcc, 0x4b, 0xc9, 0xcf, 0x75, 0x49, 0x2c, 0x49, 0x74, 0x2c, 0x4a, 0x57, 0x92, 0xe3,
	0xe2, 0x42, 0x08, 0x08, 0x09, 0x70, 0x31, 0x17, 0x97, 0x14, 0x49, 0x30, 0x2a, 0x30, 0x6a, 0x70,
	0x06, 0x81, 0x98, 0x46, 0xee, 0x5c, 0xac, 0xce, 0x20, 0x43, 0x84, 0xec, 0xb8, 0x84, 0xdd, 0x53,
	0x4b, 0x10, 0x6a, 0x83, 0x4b, 0x8a, 0x52, 0x13, 0x73, 0x85, 0x84, 0xf4, 0x12, 0x0b, 0x32, 0xf5,
	0x50, 0xcc, 0x94, 0xe2, 0x47, 0x13, 0x53, 0x62, 0x30, 0x60, 0x4c, 0x62, 0x03, 0xbb, 0xc2, 0x18,
	0x10, 0x00, 0x00, 0xff, 0xff, 0x02, 0x6c, 0x5d, 0x02, 0x94, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CacheClient is the client API for Cache service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CacheClient interface {
	GetRandomDataStream(ctx context.Context, in *RandomDataArg, opts ...grpc.CallOption) (Cache_GetRandomDataStreamClient, error)
}

type cacheClient struct {
	cc *grpc.ClientConn
}

func NewCacheClient(cc *grpc.ClientConn) CacheClient {
	return &cacheClient{cc}
}

func (c *cacheClient) GetRandomDataStream(ctx context.Context, in *RandomDataArg, opts ...grpc.CallOption) (Cache_GetRandomDataStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Cache_serviceDesc.Streams[0], "/api.Cache/GetRandomDataStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &cacheGetRandomDataStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Cache_GetRandomDataStreamClient interface {
	Recv() (*RandomData, error)
	grpc.ClientStream
}

type cacheGetRandomDataStreamClient struct {
	grpc.ClientStream
}

func (x *cacheGetRandomDataStreamClient) Recv() (*RandomData, error) {
	m := new(RandomData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CacheServer is the server API for Cache service.
type CacheServer interface {
	GetRandomDataStream(*RandomDataArg, Cache_GetRandomDataStreamServer) error
}

// UnimplementedCacheServer can be embedded to have forward compatible implementations.
type UnimplementedCacheServer struct {
}

func (*UnimplementedCacheServer) GetRandomDataStream(req *RandomDataArg, srv Cache_GetRandomDataStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method GetRandomDataStream not implemented")
}

func RegisterCacheServer(s *grpc.Server, srv CacheServer) {
	s.RegisterService(&_Cache_serviceDesc, srv)
}

func _Cache_GetRandomDataStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(RandomDataArg)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CacheServer).GetRandomDataStream(m, &cacheGetRandomDataStreamServer{stream})
}

type Cache_GetRandomDataStreamServer interface {
	Send(*RandomData) error
	grpc.ServerStream
}

type cacheGetRandomDataStreamServer struct {
	grpc.ServerStream
}

func (x *cacheGetRandomDataStreamServer) Send(m *RandomData) error {
	return x.ServerStream.SendMsg(m)
}

var _Cache_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.Cache",
	HandlerType: (*CacheServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetRandomDataStream",
			Handler:       _Cache_GetRandomDataStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cache.proto",
}
