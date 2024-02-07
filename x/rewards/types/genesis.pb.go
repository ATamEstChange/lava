// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lavanet/lava/rewards/genesis.proto

package types

import (
	fmt "fmt"
	types1 "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	types "github.com/lavanet/lava/x/timerstore/types"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the rewards module's genesis state.
type GenesisState struct {
	Params             Params             `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	RefillRewardsTS    types.GenesisState `protobuf:"bytes,2,opt,name=refillRewardsTS,proto3" json:"refillRewardsTS"`
	BasePays           []BasePayGenesis   `protobuf:"bytes,3,rep,name=base_pays,json=basePays,proto3" json:"base_pays"`
	IprpcSubscriptions []string           `protobuf:"bytes,4,rep,name=iprpc_subscriptions,json=iprpcSubscriptions,proto3" json:"iprpc_subscriptions,omitempty"`
	MinIprpcCost       types1.Coin        `protobuf:"bytes,5,opt,name=min_iprpc_cost,json=minIprpcCost,proto3" json:"min_iprpc_cost"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_02c24f4df31ca14e, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetRefillRewardsTS() types.GenesisState {
	if m != nil {
		return m.RefillRewardsTS
	}
	return types.GenesisState{}
}

func (m *GenesisState) GetBasePays() []BasePayGenesis {
	if m != nil {
		return m.BasePays
	}
	return nil
}

func (m *GenesisState) GetIprpcSubscriptions() []string {
	if m != nil {
		return m.IprpcSubscriptions
	}
	return nil
}

func (m *GenesisState) GetMinIprpcCost() types1.Coin {
	if m != nil {
		return m.MinIprpcCost
	}
	return types1.Coin{}
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "lavanet.lava.rewards.GenesisState")
}

func init() {
	proto.RegisterFile("lavanet/lava/rewards/genesis.proto", fileDescriptor_02c24f4df31ca14e)
}

var fileDescriptor_02c24f4df31ca14e = []byte{
	// 378 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x41, 0x4f, 0x22, 0x31,
	0x14, 0xc7, 0x67, 0x80, 0x25, 0xcb, 0x40, 0x76, 0x93, 0x59, 0x0e, 0xb3, 0xc4, 0x8c, 0x88, 0x1a,
	0x39, 0x75, 0x02, 0xc6, 0x8b, 0x37, 0x21, 0x86, 0x78, 0x23, 0xa0, 0x17, 0x2f, 0xa4, 0x33, 0xd6,
	0xb1, 0x09, 0xd3, 0x36, 0x7d, 0x05, 0xe5, 0x5b, 0xf8, 0xb1, 0x38, 0x72, 0xf4, 0x64, 0x0c, 0x5c,
	0xfc, 0x18, 0x66, 0xa6, 0x55, 0xc1, 0xcc, 0xa9, 0x4d, 0xfb, 0x7b, 0xbf, 0xbe, 0x7f, 0x9f, 0xd3,
	0x9a, 0xe2, 0x39, 0x66, 0x44, 0x05, 0xe9, 0x1a, 0x48, 0xf2, 0x88, 0xe5, 0x1d, 0x04, 0x31, 0x61,
	0x04, 0x28, 0x20, 0x21, 0xb9, 0xe2, 0x6e, 0xdd, 0x30, 0x28, 0x5d, 0x91, 0x61, 0x1a, 0xf5, 0x98,
	0xc7, 0x3c, 0x03, 0x82, 0x74, 0xa7, 0xd9, 0xc6, 0x41, 0xae, 0x4f, 0x60, 0x89, 0x13, 0xa3, 0x6b,
	0x1c, 0xe6, 0x22, 0x21, 0x06, 0x32, 0x11, 0x78, 0x91, 0x0b, 0x29, 0x9a, 0x10, 0x09, 0x8a, 0x4b,
	0xa2, 0xb7, 0x06, 0xf2, 0x23, 0x0e, 0x09, 0xd7, 0xb5, 0xc1, 0xbc, 0x13, 0x12, 0x85, 0x3b, 0x41,
	0xc4, 0x29, 0xd3, 0xf7, 0xad, 0xf7, 0x82, 0x53, 0x1b, 0xe8, 0x28, 0x63, 0x85, 0x15, 0x71, 0xcf,
	0x9d, 0xb2, 0x6e, 0xc5, 0xb3, 0x9b, 0x76, 0xbb, 0xda, 0xdd, 0x43, 0x79, 0xd1, 0xd0, 0x30, 0x63,
	0x7a, 0xa5, 0xe5, 0xeb, 0xbe, 0x35, 0x32, 0x15, 0xee, 0x8d, 0xf3, 0x57, 0x92, 0x7b, 0x3a, 0x9d,
	0x8e, 0x34, 0x75, 0x3d, 0xf6, 0x0a, 0x99, 0xe4, 0x78, 0x57, 0xf2, 0xdd, 0x2b, 0xda, 0x7e, 0xdb,
	0xd8, 0x7e, 0x3a, 0xdc, 0x81, 0x53, 0xf9, 0x8c, 0x0e, 0x5e, 0xb1, 0x59, 0x6c, 0x57, 0xbb, 0x47,
	0xf9, 0x5d, 0xf5, 0x30, 0x90, 0x21, 0x5e, 0x18, 0xa9, 0xf1, 0xfd, 0x0e, 0xf5, 0x29, 0xb8, 0x67,
	0xce, 0x3f, 0x2a, 0xa4, 0x88, 0x26, 0x30, 0x0b, 0x21, 0x92, 0x54, 0x28, 0xca, 0x19, 0x78, 0xa5,
	0x66, 0xb1, 0x5d, 0x31, 0xb0, 0x9b, 0x01, 0xe3, 0xed, 0x7b, 0xf7, 0xd2, 0xf9, 0x93, 0x50, 0x36,
	0xd1, 0xa5, 0x11, 0x07, 0xe5, 0xfd, 0xca, 0x52, 0xfd, 0x47, 0xfa, 0x73, 0x51, 0xfa, 0x00, 0x32,
	0x9f, 0x8b, 0xfa, 0x9c, 0x32, 0x23, 0xab, 0x25, 0x94, 0x5d, 0xa5, 0x55, 0x7d, 0x0e, 0xaa, 0x77,
	0xb1, 0x5c, 0xfb, 0xf6, 0x6a, 0xed, 0xdb, 0x6f, 0x6b, 0xdf, 0x7e, 0xde, 0xf8, 0xd6, 0x6a, 0xe3,
	0x5b, 0x2f, 0x1b, 0xdf, 0xba, 0x3d, 0x89, 0xa9, 0x7a, 0x98, 0x85, 0x28, 0xe2, 0x49, 0xb0, 0x33,
	0xd4, 0xa7, 0xaf, 0xd9, 0xab, 0x85, 0x20, 0x10, 0x96, 0xb3, 0xa1, 0x9d, 0x7e, 0x04, 0x00, 0x00,
	0xff, 0xff, 0xb8, 0x60, 0xc4, 0x73, 0x93, 0x02, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.MinIprpcCost.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if len(m.IprpcSubscriptions) > 0 {
		for iNdEx := len(m.IprpcSubscriptions) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.IprpcSubscriptions[iNdEx])
			copy(dAtA[i:], m.IprpcSubscriptions[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.IprpcSubscriptions[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.BasePays) > 0 {
		for iNdEx := len(m.BasePays) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BasePays[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	{
		size, err := m.RefillRewardsTS.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.RefillRewardsTS.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.BasePays) > 0 {
		for _, e := range m.BasePays {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.IprpcSubscriptions) > 0 {
		for _, s := range m.IprpcSubscriptions {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.MinIprpcCost.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefillRewardsTS", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RefillRewardsTS.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BasePays", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BasePays = append(m.BasePays, BasePayGenesis{})
			if err := m.BasePays[len(m.BasePays)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IprpcSubscriptions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IprpcSubscriptions = append(m.IprpcSubscriptions, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinIprpcCost", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinIprpcCost.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
