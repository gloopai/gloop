/*
 * @Author: EvanQi acheqi@gmail.com
 * @Date: 2022-07-23 12:54:58
 * @LastEditors: EvanQi
 * @LastEditTime: 2022-07-23 13:08:21
 * @Description:
 */
package lib

import (
	"fmt"
	"time"
)

type gtime struct{}

var GTime gtime

/**
 * @description: 时间戳格式化成 2006-01-02 15:04:05 的格式
 * @param {int} timeSpan
 * @return {*}
 */
func (g *gtime) TimeSpanFormat(timeSpan int) string {
	titem := time.Unix(int64(timeSpan), 0)
	return titem.Format("2006-01-02 15:04:05")
}

/**
 * @description: 文本转时间戳
 * @param {string} src
 * @param {string} 输入时间格式需要输入，例 2006-01-02 15:04:05
 * @return {*}
 */
func (g *gtime) StringToTimeSpan(src string, format string) (int64, error) {
	if src == "" {
		return 0, fmt.Errorf("原始时间文本需要输入 src")
	}
	if format == "" {
		return 0, fmt.Errorf("输入时间格式需要输入，例 2006-01-02 15:04:05")
	}
	timeLayout := format                                     //转化所需模板
	loc, _ := time.LoadLocation("Local")                     //获取时区
	theTime, _ := time.ParseInLocation(timeLayout, src, loc) //使用模板在对应时区转化为time.time类型
	return theTime.Unix(), nil
}

/**
 * @description: 文本转时间戳Int
 * @param {string} src
 * @param {string} format
 * @return {*}
 */
func (g *gtime) StringToTimeSpanInt(src string, format string) (int, error) {
	re, err := g.StringToTimeSpan(src, format)
	if err != nil {
		return 0, err
	}
	return int(re), nil
}
