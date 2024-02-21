package globalutils

import (
	"context"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalredis"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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

//针对redis get时的处理判断
func JudgeGet(err error) int {
	if err != nil {
		if err != redis.Nil {
			return globalconstants.JudgeGetError
		}
		return globalconstants.JudgeGetNil
	}
	return globalconstants.JudgeGetRes
}

//时间控制
func CompareTimeIsExpire(current, pass time.Time, limit time.Duration) bool {
	duration := current.Sub(pass)
	return duration > limit
}

// `str` 原串 | `inset` 插入串 | `index` 插入点 |
func InsertString(str, insert string, index int) string {
	if index < 0 || index >= len(str) {
		globallogger.Log.Errorln("<InsertString>: please check param with |index| or |str|")
		return ""
	}
	pre, tail := str[:index], str[index:]
	return pre + insert + tail
}

// redis scan
func AllKeyScan(ctx context.Context, patter string) []string {
	var (
		cursor uint64
		keys []string
		res  []string
		err error
	)
	for {
		res, cursor, err = globalredis.RedisCache.Scan(ctx, cursor, "patter", 100).Result()
		if err != nil {
			globallogger.Log.Errorln("<AllKeyScan> has error", err)
		}
		keys = append(keys, res...)
		// 如果 cursor 为0，说明遍历完成
		if cursor == 0 {
			break
		}
	}
	return keys
}
