package packets

type ErrorCode uint8

const (
	Success          ErrorCode = 0x00
	UnrecognizedVer  ErrorCode = 0x01
	UnrecognizedType ErrorCode = 0x02
	ErrorFormat      ErrorCode = 0x03
	ConnFailed       ErrorCode = 0x04
	NoResponse       ErrorCode = 0x05
	UnsupportedEnc   ErrorCode = 0x06
	DecFailed        ErrorCode = 0x07
	PwdLenErr        ErrorCode = 0x08
	InvChar          ErrorCode = 0x09
	InvVer           ErrorCode = 0x0A
	IsfGatewayRes    ErrorCode = 0x0B
	IsfDeviceRes     ErrorCode = 0x0C
	TunnelDisc       ErrorCode = 0x0D
	NoNeighbor       ErrorCode = 0x0E
	Busy             ErrorCode = 0x0F
	DeviceNoResp     ErrorCode = 0x10
	DeviceDisc       ErrorCode = 0x11
	EstConnFailed    ErrorCode = 0x12
	InvUUID          ErrorCode = 0x13
	NoReadPer        ErrorCode = 0x14
	NoWritePer       ErrorCode = 0x15
	NoChar           ErrorCode = 0x16
	Timeout          ErrorCode = 0x17
	Other            ErrorCode = 0x18
	InvNeighbor      ErrorCode = 0x19
	Unknown          ErrorCode = 0xff
)

type Result string
type Language string

const (
	English Language = "en-US"
	Chinese Language = "zh-CN"
)

func GetResult(code ErrorCode, language Language) Result {
	var result Result
	if language == Chinese {
		switch code {
		case Success:
			result = "成功"
		case UnrecognizedVer:
			result = "不识别的版本"
		case UnrecognizedType:
			result = "不识别的消息类型"
		case ErrorFormat:
			result = "消息格式错误"
		case ConnFailed:
			result = "连接失败"
		case NoResponse:
			result = "对端无应答"
		case UnsupportedEnc:
			result = "不支持的加密类型"
		case DecFailed:
			result = "解密错误"
		case PwdLenErr:
			result = "密码长度错误"
		case InvChar:
			result = "非法字符"
		case InvVer:
			result = "版本文件非法"
		case IsfGatewayRes:
			result = "网关资源不足"
		case IsfDeviceRes:
			result = "对端资源不足"
		case TunnelDisc:
			result = "AP隧道中断连接"
		case NoNeighbor:
			result = "无此邻居"
		case Busy:
			result = "当前正忙"
		case DeviceNoResp:
			result = "连接对接无响应"
		case DeviceDisc:
			result = "对端终止连接"
		case EstConnFailed:
			result = "建立连接失败"
		case InvUUID:
			result = "无效UUID"
		case NoReadPer:
			result = "没有读权限"
		case NoWritePer:
			result = "没有写权限"
		case NoChar:
			result = "没有该属性"
		case Timeout:
			result = "读超时或写超时"
		case Other:
			result = "其他原因造成的执行命令失败"
		case InvNeighbor:
			result = "非法邻居"
		case Unknown:
			result = "未知错误"
		}
	} else if language == English {
		switch code {
		case Success:
			result = "success"
		case UnrecognizedVer:
			result = "unrecognized version"
		case UnrecognizedType:
			result = "unrecognized message type"
		case ErrorFormat:
			result = "error format"
		case ConnFailed:
			result = "connect failed"
		case NoResponse:
			result = "device no response"
		case UnsupportedEnc:
			result = "unsupported encryption"
		case DecFailed:
			result = "decrypt failed"
		case PwdLenErr:
			result = "password length error"
		case InvChar:
			result = "invalid characteristic"
		case InvVer:
			result = "invalid version file"
		case IsfGatewayRes:
			result = "insufficient gateway resource"
		case IsfDeviceRes:
			result = "insufficient device resource"
		case TunnelDisc:
			result = "AP tunnel disconnected"
		case NoNeighbor:
			result = "no such neighbor"
		case Busy:
			result = "busy now"
		case DeviceNoResp:
			result = "device no response"
		case DeviceDisc:
			result = "device disconnected"
		case EstConnFailed:
			result = "establish connection failed"
		case InvUUID:
			result = "invalid UUID"
		case NoReadPer:
			result = "no read permission"
		case NoWritePer:
			result = "no write permission"
		case NoChar:
			result = "no such characteristic"
		case Timeout:
			result = "read or write timeout"
		case Other:
			result = "operate failed for other reason"
		case InvNeighbor:
			result = "invalid neighbor"
		case Unknown:
			result = "unknown error"
		}
	}

	return result
}
