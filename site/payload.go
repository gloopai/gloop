package site

import (
	"encoding/json"
	"net/http"

	"github.com/gloopai/gloop/lib"
)

type RequestAuth struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
}

type RequestPayload struct {
	Auth    RequestAuth `json:"auth"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}

// Data 反序列化
func (d *RequestPayload) Unmarshal(v interface{}) error {
	return lib.Convert.InterfaceToStruct(d.Data, &v)
}

// 获取page数据
func (d *RequestPayload) UnmarshalPage(v interface{}, pageSize int) (CurrentPage int, StartNum int, er error) {
	type pageObj struct {
		Page int `json:"page"`
	}
	var page pageObj
	page.Page = 1
	err := d.Unmarshal(&page)
	if err != nil {
		page.Page = 1
	}
	startNum := (page.Page - 1) * pageSize
	return page.Page, startNum, d.Unmarshal(v)
}

func (d *RequestPayload) UnmarshalPageBySize(v interface{}, pageSize int) (CurrentPage int, StartNum int, pagesize int, er error) {
	type pageObj struct {
		Page     int `json:"page"`
		PageSize int `json:"pagesize"`
	}
	var page pageObj
	page.Page = 1
	err := d.Unmarshal(&page)
	if err != nil {
		page.Page = 1
	}
	if page.PageSize > 0 {
		pageSize = page.PageSize
	}
	startNum := (page.Page - 1) * pageSize
	return page.Page, startNum, pageSize, d.Unmarshal(v)
}

/**
 * @description: 输入参数校验,tag 配置 参考 https://github.com/go-playground/validator
 * @return {*}
 */
//

func (d *RequestPayload) Validator(v interface{}) error {
	return lib.Verification.Validator(v)
}

type ResponsePayload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ParseJSONRequest(r *http.Request, payload *RequestPayload) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(payload)
}

func WriteJSONResponse(w http.ResponseWriter, response ResponsePayload) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
