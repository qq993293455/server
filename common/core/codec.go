package core

const (
	TotalLenLen  = 4
	HeaderLenLen = 2
	HeaderLen    = TotalLenLen + HeaderLenLen
)

//func DecodeRequest(data []byte) (header *models.ServerHeader, msg proto.Message, err error) {
//	// 包至少包含4字节的总长度和2字节的header长度
//	if len(data) < HeaderLen {
//		return nil, nil, ErrInvalidRequestLength
//	}
//	header = &models.Header{}
//	allLength := int(binary.LittleEndian.Uint32(data))
//	if allLength <= 0 {
//		return
//	}
//	headerLength := int(binary.LittleEndian.Uint16(data[TotalLenLen:]))
//	if headerLength > 0 {
//		headerData := data[HeaderLen : HeaderLen+headerLength]
//		err = proto.Unmarshal(headerData, header)
//		if err != nil {
//			return
//		}
//	}
//	payload = data[HeaderLen+headerLength:]
//	return
//}
//
//func EncodeResponse(header *models.Header, err *errmsg.ErrMsg, data []byte) []byte {
//	h := &models.Header{
//		TraceId:      header.TraceId,
//		ClientId:     header.ClientId,
//		ServerHeader: header.ServerHeader,
//	}
//
//	if err != nil {
//		h.Err = (*models.Error)(err)
//	}
//	hData, err1 := proto.Marshal(h)
//	MustError(err1)
//	lenHeader := len(hData)
//	lenData := len(data)
//	res := make([]byte, HeaderLen+lenHeader+lenData)
//	binary.LittleEndian.PutUint32(res, uint32(HeaderLen+lenHeader+lenData))
//	binary.LittleEndian.PutUint16(res[TotalLenLen:], uint16(lenHeader))
//	if lenHeader > 0 {
//		copy(res[HeaderLen:], hData)
//	}
//	if lenData > 0 {
//		copy(res[HeaderLen+lenHeader:], data)
//	}
//	return res
//}
