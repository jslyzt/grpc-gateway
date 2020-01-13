package runtime

import (
	"errors"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jslyzt/cast"
)

// 错误定义
var (
	ErrParamNoValue = errors.New("param has no value")
)

// Params 参数
type Params map[string]string

// String 获取string
func (pm Params) String(key string) (string, error) {
	val, ok := pm[key]
	if ok == false {
		return "", ErrParamNoValue
	}
	return val, nil
}

// StringSlice 获取[]string
func (pm Params) StringSlice(key, sep string) ([]string, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return StringSlice(val, sep)
}

// Float64 获取Float64
func (pm Params) Float64(key string) (float64, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToFloat64E(val)
}

// Float64Slice 获取[]float64
func (pm Params) Float64Slice(key, sep string) ([]float64, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Float64Slice(val, sep)
}

// Float32 获取Float32
func (pm Params) Float32(key string) (float32, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToFloat32E(val)
}

// Float32Slice 获取[]float32
func (pm Params) Float32Slice(key, sep string) ([]float32, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Float32Slice(val, sep)
}

// Int64 获取Int64
func (pm Params) Int64(key string) (int64, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToInt64E(val)
}

// Int64Slice 获取[]int64
func (pm Params) Int64Slice(key, sep string) ([]int64, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Int64Slice(val, sep)
}

// Uint64 获取Uint64
func (pm Params) Uint64(key string) (uint64, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToUint64E(val)
}

// Uint64Slice 获取[]uint64
func (pm Params) Uint64Slice(key, sep string) ([]uint64, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Uint64Slice(val, sep)
}

// Int32 获取Int32
func (pm Params) Int32(key string) (int32, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToInt32E(val)
}

// Int32Slice 获取[]int32
func (pm Params) Int32Slice(key, sep string) ([]int32, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Int32Slice(val, sep)
}

// Uint32 获取Uint32
func (pm Params) Uint32(key string) (uint32, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return cast.ToUint32E(val)
}

// Uint32Slice 获取[]uint32
func (pm Params) Uint32Slice(key, sep string) ([]uint32, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Uint32Slice(val, sep)
}

// Bool 获取bool
func (pm Params) Bool(key string) (bool, error) {
	val, ok := pm[key]
	if ok == false {
		return false, ErrParamNoValue
	}
	return cast.ToBoolE(val)
}

// BoolSlice 获取[]bool
func (pm Params) BoolSlice(key, sep string) ([]bool, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return BoolSlice(val, sep)
}

// Timestamp 获取Timestamp
func (pm Params) Timestamp(key string) (*timestamp.Timestamp, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Timestamp(val)
}

// TimestampSlice 获取[]Timestamp
func (pm Params) TimestampSlice(key, sep string) ([]*timestamp.Timestamp, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return TimestampSlice(val, sep)
}

// Duration 获取bool
func (pm Params) Duration(key string) (*duration.Duration, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Duration(val)
}

// DurationSlice 获取[]*duration.Duration
func (pm Params) DurationSlice(key, sep string) ([]*duration.Duration, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return DurationSlice(val, sep)
}

//////////////////////////////////////////////////////////////////////////////////////////

// Bytes 获取bool
func (pm Params) Bytes(key string) ([]byte, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return Bytes(val)
}

// BytesSlice 获取[][]byte
func (pm Params) BytesSlice(key, sep string) ([][]byte, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return BytesSlice(val, sep)
}

// Enum 获取enum
func (pm Params) Enum(key string, mp map[string]int32) (int32, error) {
	val, ok := pm[key]
	if ok == false {
		return 0, ErrParamNoValue
	}
	return Enum(val, mp)
}

// EnumSlice 获取[]int32
func (pm Params) EnumSlice(key, sep string, mp map[string]int32) ([]int32, error) {
	val, ok := pm[key]
	if ok == false {
		return nil, ErrParamNoValue
	}
	return EnumSlice(val, sep, mp)
}