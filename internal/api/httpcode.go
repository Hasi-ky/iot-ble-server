package api

//Http Code
const (
	HTTP_CODE_EXCEPTION = -1
	HTTP_CODE_SUCCESS   = 0
	HTTP_CODE_ERROR     = 1
)

//HTTP MESSAGE
const (
	HTTP_MESSAGE_SUCESS           = "success"
	HTTP_MESSAGE_FAILED           = "fail"
	HTTP_MEESAGE_ABSENCE_TERMINAL = "终端对应数据缺失"
	HTTP_MEESAGE_CACHE_EXCEPITON  = "缓存库异常"
	HTTP_ERROR_DATA_ABSENCE       = "部分数据缺失错误"
)

//普通错误
const (
	ERROR_DATA_FORMAT       = "数据格式错误"
	ERROR_DATA_CHANGE       = "数据转化错误"
	ERROR_CACHE_ABSENCE     = "缓存库缺少对应数据"
	ERROR_DATA_GENERATEDOWN = "下行数据生成失败"

	ERROR_CACHE_EXCEPTION = "缓存库出现异常"
)
