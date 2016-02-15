package datatype

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func AsString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func ParseStrNumberEx(src string, dest interface{}, old interface{}, op int) error {
	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errors.New("destination pointer is nil")
	}

	oldpv := reflect.ValueOf(old)
	if oldpv.Kind() != dpv.Elem().Kind() {
		return errors.New("dest and old is not kind type")
	}

	dv := reflect.Indirect(dpv)

	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		i64, err := strconv.ParseInt(src, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		switch op {
		case 1: // +
			dv.SetInt(oldpv.Int() + i64)
		case 2: // -
			dv.SetInt(oldpv.Int() - i64)
		case 3: // *
			dv.SetInt(int64(float64(oldpv.Int()) * float64(i64)))
		case 4: // /
			dv.SetInt(int64(float64(oldpv.Int()) / float64(i64)))
		case 5: // =
			dv.SetInt(i64)
		default:
			panic("unknown op")
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(src, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		switch op {
		case 1: // +
			dv.SetUint(oldpv.Uint() + u64)
		case 2: // -
			dv.SetUint(oldpv.Uint() - u64)
		case 3: // *
			dv.SetUint(uint64(float64(oldpv.Uint()) * float64(u64)))
		case 4: // /
			dv.SetUint(uint64(float64(oldpv.Uint()) / float64(u64)))
		case 5: // =
			dv.SetUint(u64)
		default:
			panic("unknown op")
		}
		return nil
	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(src, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		switch op {
		case 1: // +
			dv.SetFloat(oldpv.Float() + f64)
		case 2: // -
			dv.SetFloat(oldpv.Float() - f64)
		case 3: // *
			dv.SetFloat(oldpv.Float() * f64)
		case 4: // /
			dv.SetFloat(oldpv.Float() / f64)
		case 5: // =
			dv.SetFloat(f64)
		default:
			panic("unknown op")
		}
		return nil
	}
	return nil
}

func ParseStrNumber(src string, dest interface{}) error {
	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errors.New("destination pointer is nil")
	}

	dv := reflect.Indirect(dpv)

	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		i64, err := strconv.ParseInt(src, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(src, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(src, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	}

	return nil
}

func IsEqual(f1, f2 float64) bool {
	return math.Dim(f1, f2) < 0.00001
}

func CompareNumber(src string, dest interface{}) (int, error) {
	dv := reflect.ValueOf(dest)
	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		i64, err := strconv.ParseInt(src, 10, dv.Type().Bits())
		if err != nil {
			return 0, fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		val := dv.Int()
		switch {
		case i64 > val:
			return 1, nil
		case i64 < val:
			return -1, nil
		default:
			return 0, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		u64, err := strconv.ParseUint(src, 10, dv.Type().Bits())
		if err != nil {
			return 0, fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		val := dv.Uint()
		switch {
		case u64 > val:
			return 1, nil
		case u64 < val:
			return -1, nil
		default:
			return 0, nil
		}
	case reflect.Float32, reflect.Float64:

		f64, err := strconv.ParseFloat(src, dv.Type().Bits())
		if err != nil {
			return 0, fmt.Errorf("converting string %q to a %s: %v", src, dv.Kind(), err)
		}
		val := dv.Float()
		switch {
		case IsEqual(f64, val):
			return 0, nil
		case f64 > val:
			return 1, nil
		default:
			return -1, nil
		}
	}

	return 0, fmt.Errorf("type not match")

}

func ParseNumber(src, dest interface{}) error {
	sv := reflect.ValueOf(src)
	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errors.New("destination pointer is nil")
	}

	dv := reflect.Indirect(dpv)
	if dv.Kind() == sv.Kind() {
		dv.Set(sv)
		return nil
	}
	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := AsString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := AsString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := AsString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			return fmt.Errorf("converting string %q to a %s: %v", s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	}

	return nil
}
