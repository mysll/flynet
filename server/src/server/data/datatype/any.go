package datatype

//用于解决rpc解码过程中，不能用interface{}来解码的问题
type Any struct {
	Typ int8
	Val interface{}
}

func NewAny(t int8, v interface{}) Any {
	return Any{Typ: t, Val: v}
}

func (a *Any) Type() int8 {
	if a.Typ == DT_NONE {
		switch a.Val.(type) {
		case int8, *int8:
			a.Typ = DT_INT8
		case int16, *int16:
			a.Typ = DT_INT16
		case int32, *int32:
			a.Typ = DT_INT32
		case int64, *int64:
			a.Typ = DT_INT64
		case float32, *float32:
			a.Typ = DT_FLOAT32
		case float64, *float64:
			a.Typ = DT_FLOAT64
		case string, *string:
			a.Typ = DT_STRING
		case ObjectID, *ObjectID:
			a.Typ = DT_OBJECTID
		default:
			a.Typ = DT_INTERFACE
		}
	}
	return a.Typ
}

func (a *Any) Value() interface{} {
	return a.Val
}
