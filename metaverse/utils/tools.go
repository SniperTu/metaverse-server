package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"reflect"
)

// 转换结构体到map
func Struct2Map(v interface{}) map[string]interface{} {
	keys := reflect.TypeOf(v)
	vals := reflect.ValueOf(v)
	var rs = make(map[string]interface{})
	var length = keys.NumField()
	for i := 0; i < length; i++ {
		key := keys.Field(i).Tag.Get("bson")
		val := vals.Field(i).Interface()
		switch t := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint32, uint64, float32, float64:
			if t == 0 {
				continue
			}
		case string:
			if t == "" {
				continue
			}

		}

		rs[key] = val

	}
	return rs
}

// md5加密
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 密码加密
func Password(str string) string {
	str = Md5(str)
	str = Md5(str + "yuanyuzhou")
	return str
}

// 获取两个坐标的距离
func DistanceByPoint(a, b [2]float64) float64 {
	xl := a[0] - b[0]
	yl := a[1] - b[1]
	return math.Sqrt(math.Pow(xl, 2) + math.Pow(yl, 2))
}
