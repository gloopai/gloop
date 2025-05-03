package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gloopai/gloop/lib"
)

type RequestPayload struct {
	Auth    RequestAuth `json:"auth"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}
type RequestAuth struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
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

var Response ResponsePayload

// 返回异常
func (r *ResponsePayload) Error(msg string) ResponsePayload {
	return ResponsePayload{
		Code:    50000,
		Message: msg,
		Data:    "",
	}
}

func (r *ResponsePayload) LoginDated() ResponsePayload {
	return ResponsePayload{
		Code:    40000,
		Message: "登录过期，需要重新登录",
		Data:    nil,
	}
}

// 返回20000
func (r *ResponsePayload) SuccessNone() ResponsePayload {
	return ResponsePayload{
		Code:    20000,
		Message: "",
		Data:    nil,
	}
}

func (r *ResponsePayload) Success(v interface{}) ResponsePayload {
	return ResponsePayload{
		Code:    20000,
		Message: "",
		Data:    v,
	}
}

// 返回异常
func (r *ResponsePayload) LogError(msg string, logId string) ResponsePayload {
	return ResponsePayload{
		Code:    50000,
		Message: msg,
		Data:    "",
	}
}

// 返回列表数据
func (r *ResponsePayload) OrginList(list interface{}, page int, pagesize int, total int) ResponsePayload {
	data := make(map[string]interface{})
	data["page"] = page
	data["pagesize"] = pagesize
	data["total"] = total
	data["list"] = list
	resMap := make(map[string]interface{})
	resMap["items"] = data
	return ResponsePayload{
		Code:    20000,
		Message: "",
		Data:    resMap,
	}
}

// 返回列表数据
func (r *ResponsePayload) List(list interface{}, page int, pagesize int, total int) ResponsePayload {
	data := make(map[string]interface{})
	data["page"] = page
	data["pagesize"] = pagesize
	data["total"] = total
	format := make(map[string]interface{})
	format["create_time"] = "create_time"
	format["update_time"] = "update_time"
	resList, _ := r.ListFormatCreateTimeAndUpdateTime(list, format)
	data["list"] = resList
	return ResponsePayload{
		Code:    20000,
		Message: "",
		Data:    data,
	}
}

func (r *ResponsePayload) ListFormatCreateTimeAndUpdateTime(list interface{}, formatMap map[string]interface{}) ([]map[string]interface{}, error) {
	jsonByte, _ := json.Marshal(list)
	listMap := make([]map[string]interface{}, 0)
	err := json.Unmarshal(jsonByte, &listMap)
	if err != nil {
		return listMap, err
	}

	resMap := make([]map[string]interface{}, 0)
	for i := 0; i < len(listMap); i++ {
		itemMap := make(map[string]interface{})
		for k, v := range listMap[i] {
			formatItem, has := formatMap[k]
			if has {
				itemMap[k] = v
				timespan, _ := strconv.Atoi(fmt.Sprintf("%0.f", v))
				titem := time.Unix(int64(timespan), 0)
				itemMap[fmt.Sprintf("%v", formatItem)] = titem.Format("2006-01-02 15:04:05")
			} else {
				itemMap[k] = v
			}

		}
		resMap = append(resMap, itemMap)
	}
	return resMap, nil
}

// 返回列表数据并对指定的字段进行时间戳格式化
func (r *ResponsePayload) ListFormatDate(list interface{}, page int, pagesize int, total int, formatMap map[string]interface{}) ResponsePayload {
	resMap, _ := r.ListFormatCreateTimeAndUpdateTime(list, formatMap)
	return r.List(resMap, page, pagesize, total)
}
