package api

type Result struct {
	Code    int     `json:"code"`    //目前皆为0
	Message *string `json:"message"` //报错信息
}



type DeleteResponse Result

func (q DeleteResponse) Type() string {
	return "删除"
}

type ConnectResponse struct {
	Result
	ConnectStatus uint8 `json:"connectStatus"`
}

func (c ConnectResponse) Type() string {
	return "连接"
}

//
type MainServiceResponse struct {
	Result

}