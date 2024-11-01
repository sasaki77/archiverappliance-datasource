package pb

// Scalar Data
type NumericSamepleData interface {
	GetSecondsintoyear() uint32
	GetNano() uint32
	GetValAsFloat64() float64
}

func (x *ScalarByte) GetValAsFloat64() float64 {
	b := x.GetVal()
	if len(b) > 0 {
		return float64(x.GetVal()[0])
	}

	return 0.0
}

func (x *ScalarShort) GetValAsFloat64() float64 {
	return float64(x.GetVal())
}

func (x *ScalarInt) GetValAsFloat64() float64 {
	return float64(x.GetVal())
}

func (x *ScalarEnum) GetValAsFloat64() float64 {
	return float64(x.GetVal())
}

func (x *ScalarFloat) GetValAsFloat64() float64 {
	return float64(x.GetVal())
}

func (x *ScalarDouble) GetValAsFloat64() float64 {
	return float64(x.GetVal())
}

// Array Data
type ArraySamepleData interface {
	GetSecondsintoyear() uint32
	GetNano() uint32
	GetValAsFloat64() []float64
}

func (x *VectorChar) GetValAsFloat64() []float64 {
	v := x.GetVal()
	return convToFloat64Array(v)
}

func (x *VectorShort) GetValAsFloat64() []float64 {
	v := x.GetVal()
	return convToFloat64Array(v)
}

func (x *VectorInt) GetValAsFloat64() []float64 {
	v := x.GetVal()
	return convToFloat64Array(v)
}

func (x *VectorEnum) GetValAsFloat64() []float64 {
	v := x.GetVal()
	return convToFloat64Array(v)
}

func (x *VectorFloat) GetValAsFloat64() []float64 {
	v := x.GetVal()
	return convToFloat64Array(v)
}

func (x *VectorDouble) GetValAsFloat64() []float64 {
	return x.GetVal()
}

type ConvValue interface {
	int32 | byte | float32
}

func convToFloat64Array[T ConvValue](value []T) []float64 {
	if value == nil {
		return nil
	}

	f := make([]float64, len(value))
	for i, v := range value {
		f[i] = float64(v)
	}

	return f
}
