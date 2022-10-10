package field

import (
	"time"
)

//------------------------------ 值类型转换 ------------------------------//

// Time 格式化time.Time类型字段信息
func Time(key string, val time.Time, format ...string) Field {
	var timeFmt string
	if len(format) > 0 {
		timeFmt = format[0]
	}
	return Field{Key: key, Type: TypeTime, Integer: val.UnixMicro(), String: timeFmt}
}

// Duration 格式化time.Duration类型字段信息
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: TypeDuration, Integer: int64(val)}
}

//------------------------------ 指针类型转换 ------------------------------//

// Timep 格式化*time.Time类型字段信息
func Timep(key string, valp *time.Time, format ...string) Field {
	if valp == nil {
		return nullField(key)
	}
	return Time(key, *valp, format...)
}

// Durationp 格式化*time.Duration类型字段信息
func Durationp(key string, valp *time.Duration) Field {
	if valp == nil {
		return nullField(key)
	}
	return Duration(key, *valp)
}

//------------------------------ 切片类型转换 ------------------------------//

// Times 格式化[]time.Time类型字段信息
func Times(key string, vals []time.Time, format ...string) Field {
	var timeFmt string
	if len(format) > 0 {
		timeFmt = format[0]
	}
	return Field{Key: key, Type: TypeTimes, Interface: vals, String: timeFmt}
}

// Durations 格式化[]time.Duration类型字段信息
func Durations(key string, vals []time.Duration) Field {
	return Field{Key: key, Type: TypeDurations, Interface: vals}
}
