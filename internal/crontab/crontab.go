package crontab

import (
	"context"
	"encoding/json"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/global/globalutils"
	"iot-ble-server/internal/config"
	"time"

	"github.com/coocood/freecache"
)

func Start(ctx context.Context) error {
	globallogger.Log.Infoln("Start Crontab server")
	go TerminalAging(ctx) //终端老化

	return nil
}

//终端老化
func TerminalAging(ctx context.Context) {
	globallogger.Log.Infoln("<TerminalAging> start executing terminal aging detection strategy")
	tickerAger := time.NewTicker(globalconstants.AgingTTLDuration)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tickerAger.C:
			if config.C.General.UseRedis {
				var tempDevInfo = make(map[string]string)
				allDevInfo, err := globalredis.RedisCache.HGetAll(ctx, globalconstants.BleDevInfoCachePrefix).Result()
				if err != nil {
					globallogger.Log.Errorf("<TerminalAging>, redis has broken", err)
					continue
				}
				for devMac, devInfo := range allDevInfo {
					var tempDev globalstruct.TerminalInfo
					err := json.Unmarshal([]byte(devInfo), &tempDev)
					if err != nil {
						globallogger.Log.Errorf("<TerminalAging>, devEUI:[%s] can't resolve terminal information, please check\n", devMac)
					} else {
						if !globalutils.CompareTimeIsExpire(time.Now(), tempDev.TimeStamp, globalconstants.AgingTTLDuration) {
							tempDevInfo[devMac] = devInfo //不用更新重新加入即可
						}
					}
				}
				globalredis.RedisCache.Del(ctx, globalconstants.BleDevInfoCachePrefix)
				globalredis.RedisCache.HSet(ctx, globalconstants.BleDevInfoCachePrefix, tempDevInfo) //更新
			} else {
				var tempBleFreeCache = freecache.NewCache(10 * 1024 * 1024)
				iter := globalmemo.BleFreeCacheDevInfo.NewIterator()
				for {
					var (
						tempDev globalstruct.TerminalInfo
						tempRes = iter.Next()
					)
					if tempRes == nil {
						break
					}
					err := json.Unmarshal(tempRes.Value, &tempDev)
					if err != nil {
						globallogger.Log.Errorf("<TerminalAging>, devEUI:[%s] can't resolve terminal information, please check\n", string(tempRes.Key))
					} else {
						if !globalutils.CompareTimeIsExpire(time.Now(), tempDev.TimeStamp, globalconstants.AgingTTLDuration) {
							tempBleFreeCache.Set(tempRes.Key, tempRes.Value, 0)
						}
					}
				}
				globalmemo.BleFreeCacheDevInfo = tempBleFreeCache //更新
			}
		}
	}
}
