package globalutils

import (
	"iot-ble-server/global/globalconstants"
	"strings"
)

//首位为prefix
func CreateCacheKey(args ...string) string {
	var resultKey strings.Builder
	for argIndex, tempStr := range args {
		resultKey.WriteString(tempStr)
		if argIndex != len(args) - 1 {
			resultKey.WriteString(globalconstants.CacheSeparator)
		}
	}
	return resultKey.String()
}
