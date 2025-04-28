/*
 * @Author: EvanQi acheqi@gmail.com
 * @Date: 2022-08-02 16:53:50
 * @LastEditors: acheqi@126.com
 * @LastEditTime: 2023-06-26 11:31:17
 * @Description:
 */
package lib

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-playground/validator"
)

type verification struct{}

var Verification verification
var validate *validator.Validate

func (v *verification) Email(email string) bool {
	//pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// mobile verify
func (v *verification) Mobile(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,2,3,5-8])|(18[0-9])|166|190|198|199|191|193|195|155|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}
func (v *verification) Validator(o interface{}) error {
	validate = validator.New()
	t := reflect.TypeOf(o)
	var val = reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("输入校验时传入类型错误")
	}
	fieldNum := t.NumField()
	for i := 0; i < fieldNum; i++ {
		fieldName := t.Field(i).Name
		field, ok := t.FieldByName(fieldName)
		if ok {
			vrule := field.Tag.Get("validate")
			if vrule != "" {
				value := val.Field(i).Interface()
				err := validate.Var(value, vrule)
				if err != nil {
					msg := field.Tag.Get("validate_msg")
					if msg != "" {
						return fmt.Errorf(msg)
					} else {
						return fmt.Errorf("参数校验:%s 规则:%s", fieldName, vrule)
					}
				}
			}
		}
	}
	return nil
}
