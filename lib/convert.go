/*
 * @Author: EvanQi acheqi@gmail.com
 * @Date: 2022-05-24 19:47:52
 * @LastEditors: acheqi@126.com
 * @LastEditTime: 2023-02-15 14:29:00
 * @Description:
 */
package lib

import (
	"encoding/json"
	"reflect"

	"github.com/sirupsen/logrus"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

type convert struct{}

var Convert convert

// interface 转化成json字符串
func (c *convert) Marshal(v interface{}) ([]byte, error) {
	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
		}
	}()
	return json.Marshal(v)
}

// json字符串转化成实体对象，输入强类型校验
func (c *convert) Unmarshal(data []byte, v interface{}) error {
	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
		}
	}()
	return json.Unmarshal(data, &v)
}

// json字符串转化成实体对象，忽略输入类型
func (c *convert) UnmarshalIgnoreType(data []byte, v interface{}) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	extra.RegisterFuzzyDecoders()
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	return nil
}

// interface 转化成实体对象，输入强类型校验
func (c *convert) InterfaceToStruct(data interface{}, v interface{}) error {
	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
		}
	}()
	dataByte, err := c.Marshal(data)
	if err != nil {
		return err
	}

	err = c.Unmarshal(dataByte, &v)
	if err != nil {
		return err
	}
	return nil
}

// interface 转化实体对象，忽略输入类型
func (c *convert) InterfaceToStructIgnoreType(data interface{}, v interface{}) error {
	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
		}
	}()
	dataByte, err := c.Marshal(data)
	if err != nil {
		return err
	}

	err = c.UnmarshalIgnoreType(dataByte, &v)
	if err != nil {
		return err
	}
	return nil
}

func (c *convert) Struct2MapJson(obj interface{}) map[string]interface{} {
	defer func() {
		if e := recover(); e != nil {
			logrus.Error(e)
		}
	}()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[string(t.Field(i).Tag.Get("json"))] = v.Field(i).Interface()
	}
	return data
}
