package mir

type TypeType uint

const (
	Untyped TypeType = iota
	TypeNumber
	TypeString
	TypeBoolean
	TypeArray
	TypeDynArray
	TypeStruct
)

type Type interface {
	Type() TypeType
	GetSize() *uint
}

type TypeView struct {
	Type   Type
	Slots  []Slot
	Offset uint
}

type UntypedType struct{}

func (s *UntypedType) Type() TypeType {
	return Untyped
}

func (s *UntypedType) GetSize() *uint {
	return nil
}

type NumberType struct{}

func (s *NumberType) Type() TypeType {
	return TypeNumber
}

func (s *NumberType) GetSize() *uint {
	var len uint = 1
	return &len
}

type StringType struct{}

func (s *StringType) Type() TypeType {
	return TypeString
}

func (s *StringType) GetSize() *uint {
	var len uint = 1
	return &len
}

type BooleanType struct{}

func (s *BooleanType) Type() TypeType {
	return TypeBoolean
}

func (s *BooleanType) GetSize() *uint {
	var len uint = 1
	return &len
}

type ArrayType struct {
	Inner Type
	N     uint
}

func (s *ArrayType) Type() TypeType {
	return TypeArray
}

func (s *ArrayType) GetSize() *uint {
	var len uint = *s.Inner.GetSize() * s.N
	return &len
}

func (s *ArrayType) GetInnerType() Type {
	return s.Inner
}

type DynArrayType struct {
	Inner Type
	N     uint
}

func (s *DynArrayType) Type() TypeType {
	return TypeDynArray
}

func (s *DynArrayType) GetSize() *uint {
	return nil
}

func (s *DynArrayType) GetInnerType() Type {
	return s.Inner
}

type StructType struct {
	Fields map[string]StructField
	Size   uint
}

type StructField struct {
	Type   Type
	Offset uint
}

func (s *StructType) Type() TypeType {
	return TypeStruct
}

func (s *StructType) GetSize() *uint {
	var len uint = s.Size
	return &len
}

func (s *StructType) GetField(field string) (StructField, bool) {
	theField, err := s.Fields[field]
	return theField, err
}
