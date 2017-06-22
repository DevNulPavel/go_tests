package request

const REQUEST_TYPE_GET_SERVER_STATE int = 0
const REQUEST_TYPE_SERVER_STATE_UPDATED int = 1
const REQUEST_TYPE_CLIENT_CHANGED int = 2

type Request struct {
	RequestType int                    `json:"type"`
	Params      map[string]interface{} `json:"params"`
	SubData     RequstSubData          `json:"subdata"`
}

type RequstSubData struct {
	TestValue1 string `json:"testKey1"`
	TestValue2 string `json:"testKey2"`
}

func NewRequest(requestType int) Request {
	params := make(map[string]interface{})
	subdata := RequstSubData{}
	return Request{requestType, params, subdata}
}
