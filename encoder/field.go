package encoder

import (
	"math"
	"strconv"

	"github.com/bearki/belog/v2/field"
	"github.com/bearki/belog/v2/internal/convert"
)

// appendFieldValue 追加字段值
func appendFieldValue(dst []byte, val field.Field) []byte {
	switch true {

	case field.IsValidRange(field.TypeInt, val.ValType, field.TypeInt64):
		dst = strconv.AppendInt(dst, val.Integer, 10)

	case field.IsValidRange(field.TypeUint, val.ValType, field.TypeUint64):
		dst = strconv.AppendUint(dst, uint64(val.Integer), 10)

	case val.ValType == field.TypeFloat32:
		dst = strconv.AppendFloat(dst, float64(math.Float32frombits(uint32(val.Integer))), 'E', -1, 32)

	case val.ValType == field.TypeFloat64:
		dst = strconv.AppendFloat(dst, math.Float64frombits(uint64(val.Integer)), 'E', -1, 64)

	case field.IsValidRange(field.TypeBool, val.ValType, field.TypeString):
		dst = append(dst, val.Bytes...)

	}

	return dst
}

// AppendFieldAndMsg 将字段拼接为行格式
//
// @params dst 目标切片
//
// @params message 日志消息
//
// @params val 字段列表
//
// @return 序列化后的行格式字段字符串
//
// 返回示例，反引号内为实际内容:
// `k1: v1, k2: v2, ..., message`
func AppendFieldAndMsg(dst []byte, message string, val ...field.Field) []byte {
	// 遍历所有字段
	for _, v := range val {
		// 追加字段并序列化
		dst = append(dst, v.Key...)
		dst = append(dst, `: `...)
		dst = appendFieldValue(dst, v)
		dst = append(dst, `, `...)
	}

	// 追加message内容
	dst = append(dst, convert.StringToBytes(message)...)

	// 返回组装好的数据
	return dst
}

// AppendFieldAndMsgJSON 将字段拼接为json格式
//
// @params dst 目标切片
//
// @params messageKey 消息的键名
//
// @params message 消息内容
//
// @params fieldsKey 包裹所有字段的键名
//
// @params val 字段列表
//
// @return 序列化后的JSON格式字段字符串
//
// 返回示例，反引号内为实际内容:
// `"fields": {"k1": "v1", ...}, "msg": "message"`
func AppendFieldAndMsgJSON(dst []byte, messageKey string, message string, fieldsKey string, val ...field.Field) []byte {
	// 追加字段集字段
	dst = append(dst, '"')
	dst = append(dst, fieldsKey...)
	dst = append(dst, `": {`...)
	// 是否需要追加分隔符了
	appendDelimiter := false
	// 遍历所有字段
	for _, v := range val {
		// 从第二个有效字段开始追加分隔符号
		if appendDelimiter {
			dst = append(dst, `, `...)
		}

		// 追加字段并序列化
		dst = append(dst, '"')
		dst = append(dst, v.Key...)
		dst = append(dst, `": `...)
		dst = append(dst, v.Prefix...)
		dst = appendFieldValue(dst, v)
		dst = append(dst, v.Suffix...)

		// 已经填充了一个有效字段了
		if !appendDelimiter {
			// 下一次需要追加分隔符
			appendDelimiter = true
		}
	}
	// 追加字段结束括号
	dst = append(dst, `}, `...)

	// 追加message字段及其内容
	dst = append(dst, '"')
	dst = append(dst, messageKey...)
	dst = append(dst, `": "`...)
	dst = append(dst, convert.StringToBytes(message)...)
	dst = append(dst, '"')

	// 返回组装好的数据
	return dst
}
