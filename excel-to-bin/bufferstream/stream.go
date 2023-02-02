package bufferstream

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"unsafe"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/pierrec/lz4/v4"
)

const (
	IntT             uint64 = 0
	FloatT           uint64 = 1
	StringT          uint64 = 2
	StdVectorT       uint64 = 3
	StdArrayT        uint64 = 4
	StdMapT          uint64 = 5
	StdUnorderedMapT uint64 = 5
	CustomStructT    uint64 = 6
)

type Integer interface {
	~int8 | ~uint8 | ~int | ~uint | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type BaseStream struct {
	data_ []byte
	len_  uint64
	cap_  uint64
}

func (this_ *BaseStream) grow(newSize uint64) {
	if this_.cap_ == 0 {
		this_.cap_ = 8
		this_.data_ = make([]byte, 0, 8)
	}
	if newSize > this_.cap_ {
		this_.cap_ *= 2
		for this_.cap_ < newSize {
			this_.cap_ *= 2
		}

		newData := make([]byte, this_.len_, this_.cap_)
		copy(newData, this_.data_[:this_.len_])
		this_.data_ = newData
	}
}

func (this_ *BaseStream) WriteString(str string) {
	size := uint64(len(str))
	this_.WriteType(StringT)
	this_.writeUint64(size)
	this_.grow(this_.len_ + size)
	this_.data_ = append(this_.data_, str...)
	this_.len_ += size
}

func (this_ *BaseStream) Write(data []byte, size uint64) {
	this_.grow(this_.len_ + size)
	this_.data_ = append(this_.data_, data[:size]...)
	this_.len_ += size
}

func (this_ *BaseStream) Bytes() []byte {
	return this_.data_
}

func (this_ *BaseStream) ToLZ4() ([]byte, error) {
	newBuffer := bytes.NewBuffer(nil)
	lzw := lz4.NewWriter(newBuffer)
	err := lzw.Apply(lz4.CompressionLevelOption(lz4.Level1), lz4.BlockSizeOption(lz4.Block64Kb))
	if err != nil {
		return nil, err
	}
	n, err := lzw.Write(this_.data_)
	if uint64(n) != this_.len_ {
		return nil, fmt.Errorf("lz4 write error: need write:%d,real write:%d", this_.len_, n)
	}
	err = lzw.Close()
	if err != nil {
		return nil, err
	}
	return newBuffer.Bytes(), nil
}

func (this_ *BaseStream) ToLZ4WriteFile(path string) error {
	data, err := this_.ToLZ4()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0666)
}

func (this_ *BaseStream) WriteFile(path string) error {
	return ioutil.WriteFile(path, this_.data_, 0666)
}

func (this_ *BaseStream) WriteBool(b bool) {
	var t uint64
	if b {
		t = 1
	}
	this_.WriteType(IntT)
	this_.writeUint64(t)
}

func (this_ *BaseStream) WriteInt(u64 uint64) {
	WriteInt(this_, u64)
}

func (this_ *BaseStream) WriteFloat32(f32 float32) {
	WriteFloat(this_, f32)
}

func (this_ *BaseStream) WriteFloat64(f32 float32) {
	WriteFloat(this_, f32)
}

func (this_ *BaseStream) WriteType(typ uint64) {
	this_.writeUint64(typ)
}

func (this_ *BaseStream) writeUint64(value uint64) {
	var buffer [10]byte
	var count uint64
	for true {
		var flag uint64
		if value > 0x7f {
			flag = 0x80
		}
		buffer[count] = byte(flag) | byte(value)
		count++
		if flag > 0 {
			value >>= 7
		} else {
			break
		}
	}
	this_.Write(buffer[:count], count)
}

func WriteInt[T Integer](this_ *BaseStream, t T) {
	this_.WriteType(IntT)
	this_.writeUint64(uint64(t))
}

func WriteFloat[T Float](this_ *BaseStream, t T) {
	this_.WriteType(FloatT)
	size := int(unsafe.Sizeof(t))
	this_.writeUint64(uint64(size))
	ptr := (uintptr)(unsafe.Pointer(&t))
	sh := reflect.SliceHeader{
		Data: ptr,
		Len:  size,
		Cap:  size,
	}
	sp := (*[]byte)(unsafe.Pointer(&sh))
	this_.Write(*sp, uint64(size))
}

func Write(this_ *BaseStream, v interface{}) {
	tv := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)
	if tv.Kind() == reflect.Ptr {
		tv = tv.Elem()
		vv = vv.Elem()
	}
	switch tv.Kind() {
	case reflect.Int8, reflect.Uint8,
		reflect.Int, reflect.Uint, reflect.Int16, reflect.Uint16,
		reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
		if vv.CanInt() {
			WriteInt(this_, uint64(vv.Int()))
		} else if vv.CanUint() {
			WriteInt(this_, vv.Uint())
		} else {
			panic("Write:invalid Type")
		}
	case reflect.Bool:
		this_.WriteBool(vv.Bool())
	case reflect.String:
		this_.WriteString(vv.String())
	case reflect.Array:
		l := vv.Len()
		this_.WriteType(StdArrayT)
		this_.writeUint64(uint64(l))
		for i := 0; i < l; i++ {
			Write(this_, vv.Index(i).Interface())
		}
	case reflect.Slice:
		if tv.Elem().Kind() == reflect.Uint8 {
			l := vv.Len()
			this_.WriteType(StringT)
			this_.writeUint64(uint64(l))
			this_.Write(vv.Interface().([]byte), uint64(l))
		} else {
			l := vv.Len()
			this_.WriteType(StdVectorT)
			this_.writeUint64(uint64(l))
			for i := 0; i < l; i++ {
				Write(this_, vv.Index(i).Interface())
			}
		}
	case reflect.Map:
		this_.WriteType(StdUnorderedMapT)
		l := vv.Len()
		this_.writeUint64(uint64(l))
		keys := vv.MapKeys()
		for _, v := range keys {
			Write(this_, v.Interface())
			Write(this_, vv.MapIndex(v).Interface())
		}
	case reflect.Struct:
		switch msg := v.(type) {
		case *dynamic.Message:
			this_.WriteType(CustomStructT)
			msgFields := msg.GetMessageDescriptor().GetFields()
			for _, field := range msgFields {
				fm := msg.GetField(field)
				Write(this_, fm)
			}
		default:
			this_.WriteType(CustomStructT)
			nf := vv.NumField()
			for i := 0; i < nf; i++ {
				Write(this_, vv.Field(i).Interface())
			}
		}
	}
}
