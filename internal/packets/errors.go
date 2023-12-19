package packets

type ErrorCode string

const (
	Success          ErrorCode = "0000"
	UnrecognizedVer  ErrorCode = "0001"
	UnrecognizedType ErrorCode = "0002"
	ErrorFormat      ErrorCode = "0003"
	ConnFailed       ErrorCode = "0004"
	NoResponse       ErrorCode = "0005"
	UnsupportedEnc   ErrorCode = "0006"
	DecFailed        ErrorCode = "0007"
	PwdLenErr        ErrorCode = "0008"
	InvChar          ErrorCode = "0009"
	InvVer           ErrorCode = "000A"
	IsfGatewayRes    ErrorCode = "000B"
	IsfDeviceRes     ErrorCode = "000C"
	TunnelDisc       ErrorCode = "000D"
	NoNeighbor       ErrorCode = "000E"
	Busy             ErrorCode = "000F"
	DeviceNoResp     ErrorCode = "0010"
	DeviceDisc       ErrorCode = "0011"
	EstConnFailed    ErrorCode = "0012"
	InvUUID          ErrorCode = "0013"
	NoReadPer        ErrorCode = "0014"
	NoWritePer       ErrorCode = "0015"
	NoChar           ErrorCode = "0016"
	Timeout          ErrorCode = "0017"
	Other            ErrorCode = "0018"
	InvNeighbor      ErrorCode = "0019"
	Unknown          ErrorCode = "00ff"
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
