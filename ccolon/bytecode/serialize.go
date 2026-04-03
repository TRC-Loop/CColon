package bytecode

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/TRC-Loop/ccolon/compiler"
)

var magic = [4]byte{'C', 'C', 'L', 'B'}

const formatVersion uint16 = 2

// Type tags for constants
const (
	tagNil        byte = 0
	tagInt        byte = 1
	tagFloat      byte = 2
	tagString     byte = 3
	tagFunc       byte = 4
	tagClassDef   byte = 5
	tagBigInt     byte = 6
	tagStringList byte = 7
)

// Encode serializes a FuncObject into a portable binary format.
func Encode(fn *compiler.FuncObject, langVersion string, platform string) ([]byte, error) {
	w := &writer{}
	// header
	w.writeBytes(magic[:])
	w.writeUint16(formatVersion)
	w.writeString(langVersion)
	w.writeString(platform)
	// body
	if err := w.writeFunc(fn); err != nil {
		return nil, err
	}
	return w.buf, nil
}

// Decode deserializes bytes back into a FuncObject.
// Returns the function, language version, and any error.
func Decode(data []byte) (*compiler.FuncObject, string, error) {
	r := &reader{data: data}

	// header
	m, err := r.readN(4)
	if err != nil || m[0] != 'C' || m[1] != 'C' || m[2] != 'L' || m[3] != 'B' {
		return nil, "", fmt.Errorf("not a valid .cclb file")
	}
	fmtVer, err := r.readUint16()
	if err != nil {
		return nil, "", err
	}
	if fmtVer != formatVersion {
		return nil, "", fmt.Errorf("unsupported bytecode format version %d (expected %d)", fmtVer, formatVersion)
	}
	langVer, err := r.readString()
	if err != nil {
		return nil, "", err
	}
	_, err = r.readString() // platform (reserved)
	if err != nil {
		return nil, "", err
	}

	fn, err := r.readFunc()
	if err != nil {
		return nil, "", err
	}
	return fn, langVer, nil
}

// --- Writer ---

type writer struct {
	buf []byte
}

func (w *writer) writeBytes(b []byte) {
	w.buf = append(w.buf, b...)
}

func (w *writer) writeByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *writer) writeUint16(v uint16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	w.writeBytes(b)
}

func (w *writer) writeUint32(v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	w.writeBytes(b)
}

func (w *writer) writeInt64(v int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	w.writeBytes(b)
}

func (w *writer) writeFloat64(v float64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	w.writeBytes(b)
}

func (w *writer) writeString(s string) {
	w.writeUint32(uint32(len(s)))
	w.writeBytes([]byte(s))
}

func (w *writer) writeFunc(fn *compiler.FuncObject) error {
	w.writeString(fn.Name)
	w.writeUint16(uint16(fn.Arity))
	w.writeUint16(uint16(fn.MaxArity))
	w.writeUint16(uint16(fn.LocalCount))

	// param names
	w.writeUint16(uint16(len(fn.ParamNames)))
	for _, n := range fn.ParamNames {
		w.writeString(n)
	}

	// defaults
	w.writeUint16(uint16(len(fn.Defaults)))
	for _, d := range fn.Defaults {
		if err := w.writeConstant(d); err != nil {
			return err
		}
	}

	// code
	w.writeUint32(uint32(len(fn.Code)))
	w.writeBytes(fn.Code)

	// lines
	w.writeUint32(uint32(len(fn.Lines)))
	for _, l := range fn.Lines {
		w.writeUint32(uint32(l))
	}

	// constants
	w.writeUint32(uint32(len(fn.Constants)))
	for _, c := range fn.Constants {
		if err := w.writeConstant(c); err != nil {
			return err
		}
	}

	return nil
}

func (w *writer) writeConstant(val interface{}) error {
	switch v := val.(type) {
	case nil:
		w.writeByte(tagNil)
	case int64:
		w.writeByte(tagInt)
		w.writeInt64(v)
	case float64:
		w.writeByte(tagFloat)
		w.writeFloat64(v)
	case string:
		w.writeByte(tagString)
		w.writeString(v)
	case *compiler.FuncObject:
		w.writeByte(tagFunc)
		if err := w.writeFunc(v); err != nil {
			return err
		}
	case *compiler.ClassDef:
		w.writeByte(tagClassDef)
		if err := w.writeClassDef(v); err != nil {
			return err
		}
	case *big.Int:
		w.writeByte(tagBigInt)
		b := v.Bytes()
		w.writeByte(byte(v.Sign() + 1)) // 0=negative, 1=zero, 2=positive
		w.writeUint32(uint32(len(b)))
		w.writeBytes(b)
	case []string:
		w.writeByte(tagStringList)
		w.writeUint16(uint16(len(v)))
		for _, s := range v {
			w.writeString(s)
		}
	default:
		return fmt.Errorf("cannot serialize constant of type %T", val)
	}
	return nil
}

func (w *writer) writeClassDef(cd *compiler.ClassDef) error {
	w.writeString(cd.Name)
	w.writeString(cd.SuperName)
	w.writeUint16(uint16(cd.InitArity))
	w.writeUint16(uint16(cd.MaxArity))

	// init defaults
	w.writeUint16(uint16(len(cd.InitDefs)))
	for _, d := range cd.InitDefs {
		if err := w.writeConstant(d); err != nil {
			return err
		}
	}

	// fields
	w.writeUint16(uint16(len(cd.Fields)))
	for name, f := range cd.Fields {
		w.writeString(name)
		w.writeString(f.Visibility)
		w.writeString(f.TypeName)
		if err := w.writeConstant(f.Default); err != nil {
			return err
		}
	}

	// methods
	w.writeUint16(uint16(len(cd.Methods)))
	for name, m := range cd.Methods {
		w.writeString(name)
		w.writeString(m.Visibility)
		if err := w.writeFunc(m.Fn); err != nil {
			return err
		}
	}

	return nil
}

// --- Reader ---

type reader struct {
	data []byte
	pos  int
}

func (r *reader) readN(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("unexpected end of bytecode at offset %d", r.pos)
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *reader) readByte() (byte, error) {
	b, err := r.readN(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *reader) readUint16() (uint16, error) {
	b, err := r.readN(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (r *reader) readUint32() (uint32, error) {
	b, err := r.readN(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *reader) readInt64() (int64, error) {
	b, err := r.readN(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func (r *reader) readFloat64() (float64, error) {
	b, err := r.readN(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b)), nil
}

func (r *reader) readString() (string, error) {
	length, err := r.readUint32()
	if err != nil {
		return "", err
	}
	b, err := r.readN(int(length))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *reader) readFunc() (*compiler.FuncObject, error) {
	fn := &compiler.FuncObject{}
	var err error

	fn.Name, err = r.readString()
	if err != nil {
		return nil, err
	}

	arity, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	fn.Arity = int(arity)

	maxArity, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	fn.MaxArity = int(maxArity)

	localCount, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	fn.LocalCount = int(localCount)

	// param names
	numParams, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	fn.ParamNames = make([]string, numParams)
	for i := range fn.ParamNames {
		fn.ParamNames[i], err = r.readString()
		if err != nil {
			return nil, err
		}
	}

	// defaults
	numDefaults, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	fn.Defaults = make([]interface{}, numDefaults)
	for i := range fn.Defaults {
		fn.Defaults[i], err = r.readConstant()
		if err != nil {
			return nil, err
		}
	}

	// code
	codeLen, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	fn.Code, err = r.readN(int(codeLen))
	if err != nil {
		return nil, err
	}

	// lines
	linesLen, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	fn.Lines = make([]int, linesLen)
	for i := range fn.Lines {
		v, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		fn.Lines[i] = int(v)
	}

	// constants
	numConsts, err := r.readUint32()
	if err != nil {
		return nil, err
	}
	fn.Constants = make([]interface{}, numConsts)
	for i := range fn.Constants {
		fn.Constants[i], err = r.readConstant()
		if err != nil {
			return nil, err
		}
	}

	return fn, nil
}

func (r *reader) readConstant() (interface{}, error) {
	tag, err := r.readByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case tagNil:
		return nil, nil
	case tagInt:
		return r.readInt64()
	case tagFloat:
		return r.readFloat64()
	case tagString:
		return r.readString()
	case tagFunc:
		return r.readFunc()
	case tagClassDef:
		return r.readClassDef()
	case tagBigInt:
		sign, err := r.readByte()
		if err != nil {
			return nil, err
		}
		length, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		b, err := r.readN(int(length))
		if err != nil {
			return nil, err
		}
		v := new(big.Int).SetBytes(b)
		if sign == 0 {
			v.Neg(v)
		}
		return v, nil
	case tagStringList:
		count, err := r.readUint16()
		if err != nil {
			return nil, err
		}
		list := make([]string, count)
		for i := range list {
			list[i], err = r.readString()
			if err != nil {
				return nil, err
			}
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unknown constant tag %d at offset %d", tag, r.pos)
	}
}

func (r *reader) readClassDef() (*compiler.ClassDef, error) {
	cd := &compiler.ClassDef{
		Fields:  make(map[string]*compiler.FieldDef),
		Methods: make(map[string]*compiler.MethodDef),
	}
	var err error

	cd.Name, err = r.readString()
	if err != nil {
		return nil, err
	}
	cd.SuperName, err = r.readString()
	if err != nil {
		return nil, err
	}

	initArity, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	cd.InitArity = int(initArity)

	maxArity, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	cd.MaxArity = int(maxArity)

	// init defaults
	numDefs, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	cd.InitDefs = make([]interface{}, numDefs)
	for i := range cd.InitDefs {
		cd.InitDefs[i], err = r.readConstant()
		if err != nil {
			return nil, err
		}
	}

	// fields
	numFields, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(numFields); i++ {
		name, err := r.readString()
		if err != nil {
			return nil, err
		}
		fd := &compiler.FieldDef{}
		fd.Visibility, err = r.readString()
		if err != nil {
			return nil, err
		}
		fd.TypeName, err = r.readString()
		if err != nil {
			return nil, err
		}
		fd.Default, err = r.readConstant()
		if err != nil {
			return nil, err
		}
		cd.Fields[name] = fd
	}

	// methods
	numMethods, err := r.readUint16()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(numMethods); i++ {
		name, err := r.readString()
		if err != nil {
			return nil, err
		}
		md := &compiler.MethodDef{}
		md.Visibility, err = r.readString()
		if err != nil {
			return nil, err
		}
		md.Fn, err = r.readFunc()
		if err != nil {
			return nil, err
		}
		cd.Methods[name] = md
	}

	return cd, nil
}
