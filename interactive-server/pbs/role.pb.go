// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: role.proto

//包名

package pbs

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
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

type Role struct {
	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id" bson:"_id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name" bson:"name"`
	Gender    uint32 `protobuf:"varint,3,opt,name=gender,proto3" json:"gender" bson:"gender"`
	Avatar    string `protobuf:"bytes,7,opt,name=avatar,proto3" json:"avatar" bson:"avatar"`
	UpdatedAt int64  `protobuf:"varint,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at" bson:"updated_at"`
	DeletedAt int64  `protobuf:"varint,5,opt,name=deleted_at,json=deletedAt,proto3" json:"deleted_at" bson:"deleted_at"`
	CreatedAt int64  `protobuf:"varint,6,opt,name=created_at,json=createdAt,proto3" json:"created_at" bson:"created_at"`
}

func (m *Role) Reset()         { *m = Role{} }
func (m *Role) String() string { return proto.CompactTextString(m) }
func (*Role) ProtoMessage()    {}
func (*Role) Descriptor() ([]byte, []int) {
	return fileDescriptor_48a3ff9f7c9032f8, []int{0}
}
func (m *Role) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Role) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Role.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Role) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Role.Merge(m, src)
}
func (m *Role) XXX_Size() int {
	return m.Size()
}
func (m *Role) XXX_DiscardUnknown() {
	xxx_messageInfo_Role.DiscardUnknown(m)
}

var xxx_messageInfo_Role proto.InternalMessageInfo

func (m *Role) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Role) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Role) GetGender() uint32 {
	if m != nil {
		return m.Gender
	}
	return 0
}

func (m *Role) GetAvatar() string {
	if m != nil {
		return m.Avatar
	}
	return ""
}

func (m *Role) GetUpdatedAt() int64 {
	if m != nil {
		return m.UpdatedAt
	}
	return 0
}

func (m *Role) GetDeletedAt() int64 {
	if m != nil {
		return m.DeletedAt
	}
	return 0
}

func (m *Role) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

func init() {
	proto.RegisterType((*Role)(nil), "pbs.Role")
}

func init() { proto.RegisterFile("role.proto", fileDescriptor_48a3ff9f7c9032f8) }

var fileDescriptor_48a3ff9f7c9032f8 = []byte{
	// 340 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xbf, 0x6a, 0xeb, 0x30,
	0x1c, 0x85, 0xfd, 0x27, 0xd7, 0x17, 0x2b, 0x37, 0xc3, 0x15, 0x17, 0xae, 0x49, 0x8b, 0x14, 0xd4,
	0x25, 0xd0, 0xe2, 0xa1, 0xd9, 0xba, 0xc5, 0xa5, 0x7b, 0x51, 0x1f, 0x20, 0xd8, 0x91, 0x28, 0x86,
	0xd8, 0x32, 0x8e, 0x1a, 0xe8, 0x5b, 0x74, 0x6c, 0xdf, 0xa8, 0x63, 0xc6, 0x4e, 0xa2, 0xc4, 0x9b,
	0x47, 0x3f, 0x41, 0xb1, 0xa4, 0xa2, 0x6e, 0x3e, 0xdf, 0xe1, 0x7c, 0x16, 0xfc, 0x00, 0x68, 0xc5,
	0x8e, 0xa7, 0x4d, 0x2b, 0xa4, 0x80, 0x61, 0x53, 0xec, 0xe7, 0x7f, 0xb6, 0xa2, 0xaa, 0x44, 0x6d,
	0x10, 0x79, 0x0b, 0xc1, 0x84, 0x8a, 0x1d, 0x87, 0x29, 0x08, 0x4a, 0x96, 0xf8, 0x0b, 0x7f, 0x19,
	0x67, 0xa8, 0x57, 0x38, 0x28, 0xd9, 0xa0, 0xf0, 0xbf, 0x62, 0x2f, 0xea, 0x1b, 0xb2, 0x29, 0xd9,
	0x95, 0xa8, 0x4a, 0xc9, 0xab, 0x46, 0x3e, 0x13, 0x1a, 0x94, 0x0c, 0x5e, 0x82, 0x49, 0x9d, 0x57,
	0x3c, 0x09, 0xf4, 0xe2, 0x7f, 0xaf, 0xb0, 0xce, 0x83, 0xc2, 0x53, 0xb3, 0x19, 0x13, 0xa1, 0x1a,
	0xc2, 0x15, 0x88, 0x1e, 0x79, 0xcd, 0x78, 0x9b, 0x84, 0x0b, 0x7f, 0x39, 0xcb, 0xce, 0x7a, 0x85,
	0x2d, 0x19, 0x14, 0x9e, 0x99, 0x81, 0xc9, 0x84, 0xda, 0x62, 0x1c, 0xe5, 0x87, 0x5c, 0xe6, 0x6d,
	0xf2, 0x5b, 0xff, 0x43, 0x8f, 0x0c, 0x71, 0x23, 0x93, 0x09, 0xb5, 0x05, 0xcc, 0x00, 0x78, 0x6a,
	0x58, 0x2e, 0x39, 0xdb, 0xe4, 0x32, 0x99, 0x2c, 0xfc, 0x65, 0x98, 0x5d, 0xf4, 0x0a, 0xff, 0xa0,
	0x83, 0xc2, 0x7f, 0xcd, 0xd8, 0x31, 0x42, 0x63, 0x1b, 0xd6, 0x72, 0x74, 0x30, 0xbe, 0xe3, 0xd6,
	0xf1, 0xcb, 0x39, 0x1c, 0x75, 0x0e, 0xc7, 0x08, 0x8d, 0x6d, 0x30, 0x8e, 0x6d, 0xcb, 0xbf, 0xdf,
	0x11, 0x39, 0x87, 0xa3, 0xce, 0xe1, 0x18, 0xa1, 0xb1, 0x0d, 0x6b, 0x79, 0x9d, 0x82, 0xe9, 0x78,
	0x9a, 0x07, 0xde, 0x1e, 0xca, 0x2d, 0x87, 0x18, 0x44, 0xb7, 0xba, 0x83, 0x71, 0xda, 0x14, 0xfb,
	0x74, 0xec, 0xe6, 0x40, 0x7f, 0xde, 0x8d, 0x77, 0xc9, 0xce, 0xdf, 0x4f, 0xc8, 0x3f, 0x9e, 0x90,
	0xff, 0x79, 0x42, 0xfe, 0x4b, 0x87, 0xbc, 0xd7, 0x0e, 0x79, 0xc7, 0x0e, 0x79, 0x1f, 0x1d, 0xf2,
	0xee, 0xbd, 0x22, 0xd2, 0x27, 0x5f, 0x7d, 0x05, 0x00, 0x00, 0xff, 0xff, 0x5b, 0xb1, 0x3e, 0x64,
	0x13, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// RoleServiceClient is the client API for RoleService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type RoleServiceClient interface {
	Create(ctx context.Context, in *Role, opts ...grpc.CallOption) (*Empty, error)
}

type roleServiceClient struct {
	cc *grpc.ClientConn
}

func NewRoleServiceClient(cc *grpc.ClientConn) RoleServiceClient {
	return &roleServiceClient{cc}
}

func (c *roleServiceClient) Create(ctx context.Context, in *Role, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/pbs.RoleService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RoleServiceServer is the server API for RoleService service.
type RoleServiceServer interface {
	Create(context.Context, *Role) (*Empty, error)
}

// UnimplementedRoleServiceServer can be embedded to have forward compatible implementations.
type UnimplementedRoleServiceServer struct {
}

func (*UnimplementedRoleServiceServer) Create(ctx context.Context, req *Role) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func RegisterRoleServiceServer(s *grpc.Server, srv RoleServiceServer) {
	s.RegisterService(&_RoleService_serviceDesc, srv)
}

func _RoleService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pbs.RoleService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServiceServer).Create(ctx, req.(*Role))
	}
	return interceptor(ctx, in, info, handler)
}

var _RoleService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pbs.RoleService",
	HandlerType: (*RoleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _RoleService_Create_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "role.proto",
}

func (m *Role) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Role) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Role) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Avatar) > 0 {
		i -= len(m.Avatar)
		copy(dAtA[i:], m.Avatar)
		i = encodeVarintRole(dAtA, i, uint64(len(m.Avatar)))
		i--
		dAtA[i] = 0x3a
	}
	if m.CreatedAt != 0 {
		i = encodeVarintRole(dAtA, i, uint64(m.CreatedAt))
		i--
		dAtA[i] = 0x30
	}
	if m.DeletedAt != 0 {
		i = encodeVarintRole(dAtA, i, uint64(m.DeletedAt))
		i--
		dAtA[i] = 0x28
	}
	if m.UpdatedAt != 0 {
		i = encodeVarintRole(dAtA, i, uint64(m.UpdatedAt))
		i--
		dAtA[i] = 0x20
	}
	if m.Gender != 0 {
		i = encodeVarintRole(dAtA, i, uint64(m.Gender))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintRole(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintRole(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintRole(dAtA []byte, offset int, v uint64) int {
	offset -= sovRole(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Role) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovRole(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovRole(uint64(l))
	}
	if m.Gender != 0 {
		n += 1 + sovRole(uint64(m.Gender))
	}
	if m.UpdatedAt != 0 {
		n += 1 + sovRole(uint64(m.UpdatedAt))
	}
	if m.DeletedAt != 0 {
		n += 1 + sovRole(uint64(m.DeletedAt))
	}
	if m.CreatedAt != 0 {
		n += 1 + sovRole(uint64(m.CreatedAt))
	}
	l = len(m.Avatar)
	if l > 0 {
		n += 1 + l + sovRole(uint64(l))
	}
	return n
}

func sovRole(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozRole(x uint64) (n int) {
	return sovRole(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Role) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRole
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Role: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Role: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRole
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRole
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRole
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRole
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Gender", wireType)
			}
			m.Gender = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Gender |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdatedAt", wireType)
			}
			m.UpdatedAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdatedAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeletedAt", wireType)
			}
			m.DeletedAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DeletedAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedAt", wireType)
			}
			m.CreatedAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreatedAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Avatar", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRole
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthRole
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRole
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Avatar = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRole(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRole
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthRole
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipRole(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRole
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRole
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRole
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthRole
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupRole
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthRole
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthRole        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRole          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupRole = fmt.Errorf("proto: unexpected end of group")
)
