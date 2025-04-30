package lib

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

type generate struct{}

var Generate generate

// 当前unix 时间戳
func (g *generate) NowInt() int {
	return int(time.Now().Unix())
}

// uuid string
func (g *generate) Guid() string {
	return fmt.Sprintf("%v", uuid.New().String())
}

func (g *generate) RandNum(maxNum int) int {
	// rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Intn(maxNum)
}
func (g *generate) RandDigitCode(n int) string {
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, rand.Intn(int(math.Pow10(n))))
}

/**
 * @description: 浮点型保留小数
 * @param {float64} n
 * @param {int} decimals
 * @return {*}
 */
func (g *generate) TruncateTo(n float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Trunc(n*factor) / factor
}

// roundTo 进行四舍五入，保留指定的小数位数
func (g *generate) RoundTo(n float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(n*factor) / factor
}

// 获取中文周几
func (g *generate) WeedDayCN(t time.Time) string {
	weekday := strings.ToLower(t.Weekday().String())
	switch weekday {
	case "sunday":
		return "日"
	case "monday":
		return "一"
	case "tuesday":
		return "二"
	case "wednesday":
		return "三"
	case "thursday":
		return "四"
	case "friday":
		return "五"
	case "saturday":
		return "六"
	default:
		return "null"
	}
}

// md5
func (g *generate) Md5(str string) string {
	return Crypto.Md5(str)
}

/**
 * @description: 生成一个订单号 日期20191025时间戳1571987125435+3位随机数
 * @return {*}
 */

func (g *generate) GenerateCode() string {
	date := time.Now().Format("20060102")
	r := rand.Intn(1000)
	code := fmt.Sprintf("%s%d%03d", date, time.Now().Unix(), r)
	return code
}

func (g *generate) GenerateTradeNo() string {
	return fmt.Sprintf("%s%s", g.GenerateCode(), g.RandCode(4))
}

/**
 * @description: 时间戳格式化成 2006-01-02 15:04:05 的格式
 * @param {int} timeSpan
 * @return {*}
 */
func (g *generate) TimeSpanFormat(timeSpan int) string {
	// titem := time.Unix(int64(timeSpan), 0)
	// return titem.Format("2006-01-02 15:04:05")
	return GTime.TimeSpanFormat(timeSpan)
}

func (g *generate) RandStringBytes(n int) string {
	letterBytes := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (g *generate) RandCode(n int) string {
	letterBytes := "0123456789"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
