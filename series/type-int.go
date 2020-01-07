package series

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// intElements is the concrete implementation of Elements for Int elements.
type intElements struct {
	data []intElement
	nan  []bool
}

func (es intElements) Len() int           { return len(es.data) }
func (es intElements) Elem(i int) Element { return &es.data[i] }

type intElement struct {
	e int
}

func (es *intElement) Set(i int, value interface{}) {
	es.nan[i] = false
	switch value.(type) {
	case string:
		if value.(string) == "NaN" {
			es[i].nan = true
			return
		}
		v, err := strconv.Atoi(value.(string))
		if err != nil {
			es.data[i].nan = true
			return
		}
		es.data[i] = v
	case int:
		es.data[i] = int(value.(int))
	case float64:
		f := value.(float64)
		if math.IsNaN(f) ||
			math.IsInf(f, 0) ||
			math.IsInf(f, 1) {
			es.nan[i] = true
			return
		}
		es.data[i] = int(f)
	case bool:
		b := value.(bool)
		if b {
			es.data[i] = 1
		} else {
			es.data[i] = 0
		}
	case Element:
		v := value.(Element).ConvertTo(Int)
		if v.Type() != Int {
			es.nan[i] = true
		} else {
			es.data[i] = int(v.Int())
		}
	default:
		es.nan[i] = true
		return
	}
	return
}

func (e intElement) Copy() Element {
	if e.IsNA() {
		return &intElement{0, true}
	}
	return &intElement{e.e, false}
}

func (e intElement) IsNA() bool {
	if e.nan {
		return true
	}
	return false
}

func (e intElement) Type() reflect.Type {
	return reflect.TypeOf(e.e)
}

func (e intElement) Val() ElementValue {
	if e.IsNA() {
		return nil
	}
	return int(e.e)
}

func (e intElement) String() string {
	if e.IsNA() {
		return "NaN"
	}
	return fmt.Sprint(e.e)
}

func (e intElement) Value() reflect.Value {
	return reflect.ValueOf(e.e)
}

func (e intElement) ConvertTo(ty reflect.Type) reflect.Value {
	switch ty {
	case Bool:
		if e.IsNA() {
			return reflect.ValueOf(fmt.Errorf("can't convert NaN to %v", ty))
		}
		switch e.e {
		case 0:
			return reflect.ValueOf(false)
		case 1:
			return reflect.ValueOf(true)
		default:
			return reflect.ValueOf(fmt.Errorf("can't convert Int \"%v\" to bool", e.e))
		}
	case Int:
		if e.IsNA() {
			return reflect.ValueOf(fmt.Errorf("can't convert NaN to %v", ty))
		}
		return reflect.ValueOf(e.e)
	case Float:
		if e.IsNA() {
			return reflect.ValueOf(math.NaN())
		}
		return reflect.ValueOf(float64(e.e))
	case String:
		return reflect.ValueOf(e.String())
	default:
		return reflect.ValueOf(fmt.Errorf("unsupported type: %s", ty.String()))
	}
}
