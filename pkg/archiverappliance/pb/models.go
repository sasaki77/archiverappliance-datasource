package pb

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
