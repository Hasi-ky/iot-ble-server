package globalutils

import (
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"strconv"
	"strings"
)

//首位为prefix
func CreateCacheKey(args ...string) string {
	var resultKey strings.Builder
	for argIndex, tempStr := range args {
		resultKey.WriteString(tempStr)
		if argIndex != len(args)-1 {
			resultKey.WriteString(globalconstants.CacheSeparator)
		}
	}
	return resultKey.String()
}

//`指针偏移是否超过限制`
func JudgePacketLenthLimit(cur int, limit int) bool {
	return cur >= limit
}

// `param` 待生成数据 | `expectLen` 期望长 |
func ConvertDecimalToHexStr(param, expectLen int) string {
	hexStr := strconv.FormatInt(int64(param), 16)
	if len(hexStr) > expectLen {
		globallogger.Log.Errorln("<ConvertDecimalToHexStr>: please check param with |param| or |expectLen|")
		return ""
	}
	for len(hexStr) < expectLen {
		hexStr = "0" + hexStr
	}
	return hexStr
}
