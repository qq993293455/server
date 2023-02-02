package protocol

import (
	"encoding/binary"
	"strings"
	"unsafe"

	"coin-server/common/errmsg"
	"coin-server/common/msgcreate"
	"coin-server/common/proto/models"

	"github.com/gogo/protobuf/types"

	"github.com/gogo/protobuf/proto"
)

type MarshalTo interface {
	MarshalToSizedBuffer(dAtA []byte) (int, error)
}

const (
	totalLenLen        = 4
	protocolVersionLen = 1
	messageTypeLen     = 1
	rpcIndexLen        = 4
	msgIdLen           = 1
	serverHeaderLen    = 2

	len1      = totalLenLen + protocolVersionLen
	len2      = len1 + messageTypeLen
	len3      = len2 + rpcIndexLen
	len4      = len3 + msgIdLen
	minMsgLen = len4 + serverHeaderLen
)

type PacketV1 struct {
	TotalLen        uint32
	ProtocolVersion uint8
	MessageType     uint8
	RPCIndex        uint32
	MsgIDLen        uint8
	MsgID           string
	ServerHeaderLen uint16
	ServerHeader    *models.ServerHeader
	Msg             proto.Message
}

var errorEncodeTCP = errmsg.NewProtocolErrorInfo("data_len_too_short")

func EncodeTCPFrom(data []byte, messageType uint8, rpcIndex uint32, h *models.ServerHeader, msg proto.Message) *errmsg.ErrMsg {
	preLen := len4
	msgName := proto.MessageName(msg)
	msgNameLen := len(msgName)
	hs := 0
	if h != nil {
		hs = h.Size()
	}

	ms := proto.Size(msg)
	totalLen := preLen + msgNameLen + serverHeaderLen + hs + ms
	if len(data) != totalLen {
		return errorEncodeTCP
	}
	binary.LittleEndian.PutUint32(data, uint32(totalLen))
	data[len1] = messageType
	binary.LittleEndian.PutUint32(data[len2:], rpcIndex)
	data[len3] = uint8(msgNameLen)
	copy(data[len4:len4+msgNameLen], msgName)
	return EncodeInternalFrom(data[len4+msgNameLen:], h, msg)
}

func EncodeTCPDataFrom(data []byte, p *PacketV1, msgName string, msg []byte) *errmsg.ErrMsg {
	preLen := len4
	msgNameLen := len(msgName)
	hs := 0
	if p.ServerHeader != nil {
		hs = p.ServerHeader.Size()
	}
	ms := len(msg)
	totalLen := preLen + msgNameLen + serverHeaderLen + hs + ms
	if len(data) != totalLen {
		return errorEncodeTCP
	}
	binary.LittleEndian.PutUint32(data, uint32(totalLen))
	data[len1] = p.MessageType
	binary.LittleEndian.PutUint32(data[len2:], p.RPCIndex)
	data[len3] = uint8(msgNameLen)
	copy(data[len4:len4+msgNameLen], msgName)
	return EncodeInternalDataFrom(data[len4+msgNameLen:], p.ServerHeader, msg)
}

func EncodeTCPInternalDataFrom(data []byte, messageType uint8, rpcIndex uint32, msgName string, internalData []byte) *errmsg.ErrMsg {
	preLen := len4
	msgNameLen := len(msgName)

	totalLen := preLen + msgNameLen + len(internalData)
	if len(data) != totalLen {
		return errorEncodeTCP
	}
	binary.LittleEndian.PutUint32(data, uint32(totalLen))
	data[len1] = messageType
	binary.LittleEndian.PutUint32(data[len2:], rpcIndex)
	data[len3] = uint8(msgNameLen)
	copy(data[len4:len4+msgNameLen], msgName)
	copy(data[len4+msgNameLen:], internalData)
	return nil
}

func GetEncodeLen(h *models.ServerHeader, msg proto.Message) int {
	preLen := len4
	msgName := proto.MessageName(msg)
	msgNameLen := len(msgName)
	hs := 0
	if h != nil {
		hs = h.Size()
	}
	ms := proto.Size(msg)
	totalLen := preLen + msgNameLen + serverHeaderLen + hs + ms
	return totalLen
}

func GetEncodeDataLen(h *models.ServerHeader, msgName string, data []byte) int {
	preLen := len4
	msgNameLen := len(msgName)
	hs := 0
	if h != nil {
		hs = h.Size()
	}
	ms := len(data)
	totalLen := preLen + msgNameLen + serverHeaderLen + hs + ms
	return totalLen
}

func GetEncodeInternalDataLen(msgName string, data []byte) int {
	preLen := len4
	msgNameLen := len(msgName)
	totalLen := preLen + msgNameLen + len(data)
	return totalLen
}

const MaxMsgLen = 1 << 24

var ErrorDecodeTCP = errmsg.NewProtocolErrorInfo("decode_tcp_error")

func DecodeTCP(data []byte) (*PacketV1, *errmsg.ErrMsg) {
	ld := len(data)
	if ld < minMsgLen {
		return nil, nil
	}
	totalLen := binary.LittleEndian.Uint32(data)
	if totalLen > MaxMsgLen {
		return nil, ErrorDecodeTCP
	}
	if uint32(len(data)) < totalLen {
		return nil, nil
	}
	p := &PacketV1{TotalLen: totalLen}
	p.ProtocolVersion = data[totalLenLen]
	p.MessageType = data[len1]
	p.RPCIndex = binary.LittleEndian.Uint32(data[len2:])
	p.MsgIDLen = data[len3]
	msgIdBytes := data[len4 : len4+p.MsgIDLen]
	p.MsgID = *(*string)(unsafe.Pointer(&msgIdBytes))
	msg := msgcreate.NewMessage(p.MsgID)
	h := &models.ServerHeader{}
	err := DecodeInternal(data[len4+p.MsgIDLen:totalLen], h, msg)
	if err != nil {
		return nil, err
	}
	p.ServerHeader = h
	p.Msg = msg
	return p, nil
}

func DecodeTCPTo(data []byte, p *PacketV1) *errmsg.ErrMsg {
	ld := len(data)
	if ld < totalLenLen {
		return nil
	}
	totalLen := binary.LittleEndian.Uint32(data)
	if uint32(len(data)) < totalLen {
		return nil
	}
	p.TotalLen = totalLen
	p.ProtocolVersion = data[totalLenLen]
	p.MessageType = data[len1]
	p.RPCIndex = binary.LittleEndian.Uint32(data[len2:])
	p.MsgIDLen = data[len3]
	msgIdBytes := data[len4 : len4+p.MsgIDLen]
	p.MsgID = *(*string)(unsafe.Pointer(&msgIdBytes))
	msg := msgcreate.NewMessage(p.MsgID)
	h := &models.ServerHeader{}
	err := DecodeInternal(data[len4+p.MsgIDLen:totalLen], h, msg)
	if err != nil {
		return err
	}
	p.ServerHeader = h
	p.Msg = msg
	return nil
}

func DecodeTCPRaw(data []byte) (msgType uint8, rpcIndex uint32, msgName string, raw []byte, err *errmsg.ErrMsg) {
	ld := len(data)
	if ld < totalLenLen {
		return 0, 0, "", nil, ErrorDecodeTCP
	}
	totalLen := binary.LittleEndian.Uint32(data)
	if uint32(len(data)) < totalLen {
		return 0, 0, "", nil, ErrorDecodeTCP
	}
	msgType = data[len1]
	rpcIndex = binary.LittleEndian.Uint32(data[len2:])
	msgIDLen := data[len3]
	msgIdBytes := data[len4 : len4+msgIDLen]
	msgName = *(*string)(unsafe.Pointer(&msgIdBytes))
	raw = data[len4+msgIDLen:]
	return
}

var ErrorDecodeInternalInvalidData = errmsg.NewProtocolErrorInfo("decode_internal_error")

// DecodeInternal 解 ServerHeader+Data
func DecodeInternal(data []byte, outHeader *models.ServerHeader, out proto.Message) *errmsg.ErrMsg {
	ld := len(data)
	if ld < serverHeaderLen {
		return ErrorDecodeInternalInvalidData
	}
	hLen := int(binary.LittleEndian.Uint16(data[:serverHeaderLen]))
	if hLen > ld-serverHeaderLen {
		return ErrorDecodeInternalInvalidData
	}
	if hLen > 0 && outHeader != nil {
		err := outHeader.Unmarshal(data[serverHeaderLen : serverHeaderLen+hLen])
		if err != nil {
			return errmsg.NewProtocolError(err)
		}
	}
	return errmsg.NewProtocolError(proto.Unmarshal(data[serverHeaderLen+hLen:], out))
}

func DecodeHeaderInternal(data []byte, outHeader *models.ServerHeader) (msgData []byte, err *errmsg.ErrMsg) {
	ld := len(data)
	if ld < serverHeaderLen {
		return nil, ErrorDecodeInternalInvalidData
	}
	hLen := int(binary.LittleEndian.Uint16(data[:serverHeaderLen]))
	if hLen > ld-serverHeaderLen {
		return nil, ErrorDecodeInternalInvalidData
	}
	if hLen > 0 && outHeader != nil {
		err := outHeader.Unmarshal(data[serverHeaderLen : serverHeaderLen+hLen])
		if err != nil {
			return nil, errmsg.NewProtocolError(err)
		}
	}
	return data[serverHeaderLen+hLen:], nil
}

var ErrorEncodeInternalInvalidData = errmsg.NewProtocolErrorInfo("encode_internal_error")

func GetEncodeInternalToSize(inHeader *models.ServerHeader, in proto.Message) int {
	return serverHeaderLen + inHeader.Size() + proto.Size(in)
}

func GetEncodeInternalDataToSize(inHeader *models.ServerHeader, data []byte) int {
	return serverHeaderLen + inHeader.Size() + len(data)
}

func EncodeInternalFrom(d []byte, inHeader *models.ServerHeader, in proto.Message) *errmsg.ErrMsg {
	hs := 0
	if inHeader != nil {
		hs = inHeader.Size()
	}
	os := proto.Size(in)
	if len(d) != serverHeaderLen+hs+os {
		return ErrorEncodeInternalInvalidData
	}
	binary.LittleEndian.PutUint16(d, uint16(hs))
	if inHeader != nil {
		_, err := inHeader.MarshalToSizedBuffer(d[serverHeaderLen : serverHeaderLen+hs])
		if err != nil {
			return errmsg.NewProtocolError(err)
		}
	}
	inv := in.(MarshalTo)
	_, err := inv.MarshalToSizedBuffer(d[serverHeaderLen+hs:])
	if err != nil {
		return errmsg.NewProtocolError(err)
	}
	return nil
}

func EncodeInternalDataFrom(d []byte, inHeader *models.ServerHeader, data []byte) *errmsg.ErrMsg {
	hs := 0
	if inHeader != nil {
		hs = inHeader.Size()
	}
	os := len(data)
	if len(d) != serverHeaderLen+hs+os {
		return ErrorEncodeInternalInvalidData
	}
	binary.LittleEndian.PutUint16(d, uint16(hs))
	if inHeader != nil {
		_, err := inHeader.MarshalToSizedBuffer(d[serverHeaderLen : serverHeaderLen+hs])
		if err != nil {
			return errmsg.NewProtocolError(err)
		}
	}
	if os > 0 {
		copy(d[serverHeaderLen+hs:], data)
	}
	return nil
}

func GetMsgType(data []byte) models.MsgType {
	if len(data) < len2 {
		return models.MsgType_invalid_msg_type
	}
	return models.MsgType(data[len2-1])
}

func GetMsgDataFromRaw(raw []byte) []byte {
	hLen := binary.LittleEndian.Uint16(raw)
	return raw[serverHeaderLen+hLen:]
}

func EncodeRespToClient(msg proto.Message, msgS ...proto.Message) *models.Resp {
	r := &models.Resp{}
	r.Resp = MsgToAny(msg)
	for _, v := range msgS {
		r.OtherMsg = append(r.OtherMsg, MsgToAny(v))
	}
	return r
}

func MsgToAny(msg proto.Message) *types.Any {
	d, _ := proto.Marshal(msg)
	return &types.Any{
		TypeUrl: proto.MessageName(msg),
		Value:   d,
	}
}

func AnyToMsg(any *types.Any) (proto.Message, *errmsg.ErrMsg) {
	msg := msgcreate.NewMessage(any.TypeUrl)
	return msg, errmsg.NewProtocolError(proto.Unmarshal(any.Value, msg))
}

var (
	TopicGatewayStdTcp            = models.ServerType_GatewayStdTcp.String()
	TopicGameServer               = models.ServerType_GameServer.String()
	TopicBroadcast                = "broadcast"
	TopicBroadcastPre             = "broadcast."
	TopicBroadcastGatewayStdTcp   = strings.Join([]string{TopicBroadcast, TopicGatewayStdTcp, ">"}, ".")
	TopicBroadcastTopicGameServer = strings.Join([]string{TopicBroadcast, TopicGameServer, ">"}, ".")
)
