package api

import (
	"context"
	"errors"
	"iot-ble-server/global/globalconstants"
	"iot-ble-server/global/globallogger"
	"iot-ble-server/global/globalmemo"
	"iot-ble-server/global/globalredis"
	"iot-ble-server/global/globalstruct"
	"iot-ble-server/internal/bleudp"
	"iot-ble-server/internal/config"
	"iot-ble-server/internal/packets"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

//返回网页响应接口
//成功
func Return(ctx *gin.Context, response interface{}) {
	ctx.JSON(200, response)
}

//终端设备的连接与断开
func connectDev(ctx *gin.Context) {
	var (
		devCon        globalstruct.DevConnection
		terminal      globalstruct.TerminalInfo
		resultMessage ResultMessage = ResultMessage{
			Code: HTTP_CODE_ERROR,
		}
		err        error
		devInfoStr string
		devInfo    []byte
	)
	err = ctx.BindJSON(&devCon)
	if err != nil {
		globallogger.Log.Errorf("<connectDev>: current data conversion failed : %v\n", err)
		resultMessage.Code = HTTP_CODE_ERROR
		resultMessage.Message = err.Error() + ERROR_DATA_CHANGE
		ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
		return
	}
	//先去查缓存，缓存查到生成jsonInfo数据，走下行编码思路
	if config.C.General.UseRedis {
		devInfoStr, err = globalredis.RedisCache.HGet(ctx, globalconstants.BleDevInfoCachePrefix, devCon.DevMac).Result()
		devInfo = []byte(devInfoStr)
		if err != nil {
			if err != redis.Nil {
				globallogger.Log.Errorln("<connectDev> redis cache occur error", err)
				resultMessage.Code = HTTP_CODE_ERROR
				resultMessage.Message = ERROR_CACHE_EXCEPTION
				ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
			} else {
				resultMessage.Message = ERROR_CACHE_ABSENCE
				ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
			}
			return
		}
	} else {
		devInfo, err = globalmemo.BleFreeCacheDevInfo.Get([]byte(devCon.DevMac))
		if err != nil {
			globallogger.Log.Errorln("<connectDev> memo cache occur error", err)
			resultMessage.Code = HTTP_CODE_ERROR
			resultMessage.Message = ERROR_CACHE_EXCEPTION
			ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
			return
		}
	}
	err = json.Unmarshal([]byte(devInfo), &terminal)
	if err != nil {
		globallogger.Log.Errorf("<connectDev>: switch terminal data failed : %v\n", err)
		resultMessage.Code = HTTP_CODE_ERROR
		resultMessage.Message = ERROR_DATA_GENERATEDOWN
		ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
		return
	}
	jsonInfo, err1 := bleudp.GenerateJsonInfo(ctx, terminal, globalconstants.CtrlAllMsg, packets.BleRequest, packets.TLVConnectMsg, devCon)
	if err1 != nil {
		globallogger.Log.Errorf("<connectDev>: generate down data failed : %v\n", err)
		resultMessage.Code = HTTP_CODE_ERROR
		resultMessage.Message = err.Error()
		ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
		return
	}
	byteToSend := bleudp.EnCodeForDownUdpMessage(jsonInfo)
	curSN, _ := strconv.ParseInt(jsonInfo.MessageAppHeader.SN, 16, 16)
	bleudp.SendMsgBeforeDown(ctx, byteToSend, int(curSN), terminal.TerminalMac, terminal.GwMac)
}

// (批量)终端删除，无缺失的数据继续执行
// 入参为待删除终端的字符串数组(mac)
// 异常更加严重,如缓存崩溃
func deleteDev(ctx *gin.Context) {
	var devReq = make([]string, 0)
	var resultMessage = ResultMessage{
		Message: HTTP_MESSAGE_SUCESS,
	}
	err := ctx.BindJSON(&devReq)
	if err != nil {
		globallogger.Log.Errorf("<deleteDev>: current data conversion failed : %v\n", err)
		resultMessage.Code = HTTP_CODE_ERROR
		resultMessage.Message = err.Error() + ERROR_DATA_CHANGE
		ctx.JSON(HTTP_CODE_EXCEPTION, resultMessage)
		return
	}
	for _, waitedDelDev := range devReq {
		if waitedDelDev == "" {
			globallogger.Log.Errorln("<deleteDev>: abnormal parameter passing")
			resultMessage.Code = HTTP_CODE_ERROR
			resultMessage.Message = HTTP_ERROR_DATA_ABSENCE
			continue
		}
		err = FlushDevCacheInfo(ctx, waitedDelDev)
		if err != nil {
			if err.Error() == HTTP_MEESAGE_ABSENCE_TERMINAL {
				globallogger.Log.Errorf("<deleteDev>: absence terminal: %v\n", waitedDelDev)
				resultMessage.Code = HTTP_CODE_ERROR
				resultMessage.Message = HTTP_MEESAGE_ABSENCE_TERMINAL
			} else {
				globallogger.Log.Errorf("<deleteDev>: cache occur error %v\n", err)
				resultMessage.Code = HTTP_CODE_EXCEPTION //缓存库崩溃处理
				resultMessage.Message = HTTP_MEESAGE_CACHE_EXCEPITON
				goto final
			}
		}
	}
final:
	Return(ctx, DeleteResponse{Code: resultMessage.Code, Message: &resultMessage.Message})
}

//清除设备缓存信息
func FlushDevCacheInfo(ctx context.Context, devMac string) error {
	globallogger.Log.Infof("<FlushDevCacheInfo>: start deleting terminal: [%s] information\n", devMac)
	var err error
	if config.C.General.UseRedis {
		_, err = globalredis.RedisCache.HDel(ctx, globalconstants.BleDevInfoCachePrefix, devMac).Result()
		if err != nil {
			if err != redis.Nil {
				return errors.New("redis occur error with " + err.Error())
			}
			return errors.New(ERROR_CACHE_ABSENCE)
		}
	} else {
		finshedDel := globalmemo.BleFreeCacheDevInfo.Del([]byte(devMac))
		if !finshedDel {
			return errors.New(ERROR_CACHE_ABSENCE)
		}
	}
	globallogger.Log.Warnf("<FlushDevCacheInfo>: deleting terminal [%s] cache information\n", devMac)
	return nil
}