package text

import (
	"fmt"
	"github.com/goark/koyomi/value"
	"github.com/goark/koyomi/zodiac"
	"strconv"
	"strings"
	"time"
)

// ConvertToEraDate は、与えられた日付を和暦の日付文字列に変換します。
// 日付を受け取り、和暦の日付文字列とエラーを返します。
func ConvertToEraDate(dt time.Time) (string, error) {
	return ConvertToEraDateWithFormat(dt, "ggggyy年MM月dd日")
}

// ConvertToEraDateWithFormat は、与えられた日付を指定されたフォーマットの和暦日付文字列に変換します。
// 日付とフォーマット文字列を受け取り、和暦の日付文字列とエラーを返します。
func ConvertToEraDateWithFormat(dt time.Time, format string) (string, error) {
	te := value.NewDate(dt)
	n, y := te.YearEraString()
	if len(n) == 0 {
		return "", error(fmt.Errorf("年号が見つかりません"))
	}

	tmp := strings.ReplaceAll(y, "年", "")
	year, err := strconv.Atoi(tmp)
	if err != nil {
		return "", error(fmt.Errorf("年号が見つかりません"))
	}

	var result = format
	result = strings.ReplaceAll(result, "gggg", n)
	result = strings.ReplaceAll(result, "yy", fmt.Sprintf("%02d", year))
	result = strings.ReplaceAll(result, "MM", fmt.Sprintf("%02d", te.Month()))
	result = strings.ReplaceAll(result, "dd", fmt.Sprintf("%02d", te.Day()))
	result = strings.ReplaceAll(result, "HH", fmt.Sprintf("%02d", dt.Hour()))
	result = strings.ReplaceAll(result, "mm", fmt.Sprintf("%02d", dt.Minute()))
	result = strings.ReplaceAll(result, "ss", fmt.Sprintf("%02d", dt.Second()))

	result = strings.ReplaceAll(result, "M", fmt.Sprintf("%d", te.Month()))
	result = strings.ReplaceAll(result, "d", fmt.Sprintf("%d", te.Day()))
	result = strings.ReplaceAll(result, "H", fmt.Sprintf("%d", dt.Hour()))
	result = strings.ReplaceAll(result, "m", fmt.Sprintf("%d", dt.Minute()))
	result = strings.ReplaceAll(result, "s", fmt.Sprintf("%d", dt.Second()))

	return result, nil
}

// ConvertToZodiacByYear は、与えられた年を干支に変換します。
// 年を受け取り、干支の文字列とエラーを返します。
func ConvertToZodiacByYear(year int) (string, error) {
	kan, shi := zodiac.ZodiacYearNumber(year)
	return fmt.Sprintf("%v%v", kan, shi), nil
}

// ConvertToZodiacByDate は、与えられた日付の年を干支に変換します。
// 日付を受け取り、干支の文字列とエラーを返します。
func ConvertToZodiacByDate(dt time.Time) (string, error) {
	return ConvertToZodiacByYear(dt.Year())
}

// ConvertToDayZodiacByDate は、与えられた日付を日干支に変換します。
// 日付を受け取り、日干支の文字列とエラーを返します。
func ConvertToDayZodiacByDate(dt time.Time) (string, error) {
	t := value.NewDate(dt)
	kan, shi := zodiac.ZodiacDayNumber(t)
	return fmt.Sprintf("%v%v", kan, shi), nil
}
