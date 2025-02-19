package text

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/goark/koyomi/value"
	"github.com/goark/koyomi/zodiac"
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
		return "", fmt.Errorf("年号が見つかりません")
	}

	tmp := strings.ReplaceAll(y, "年", "")
	year, err := strconv.Atoi(tmp)
	if err != nil {
		return "", fmt.Errorf("年号が見つかりません")
	}

	rep := strings.NewReplacer(
		"gggg", n,
		"yy", fmt.Sprintf("%02d", year),
		"MM", fmt.Sprintf("%02d", te.Month()),
		"dd", fmt.Sprintf("%02d", te.Day()),
		"HH", fmt.Sprintf("%02d", dt.Hour()),
		"mm", fmt.Sprintf("%02d", dt.Minute()),
		"ss", fmt.Sprintf("%02d", dt.Second()),
		"M", fmt.Sprintf("%d", te.Month()),
		"d", fmt.Sprintf("%d", te.Day()),
		"H", fmt.Sprintf("%d", dt.Hour()),
		"m", fmt.Sprintf("%d", dt.Minute()),
		"s", fmt.Sprintf("%d", dt.Second()),
	)

	result := rep.Replace(format)

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

// CamelToKebab はキャメルケースの文字列を受け取り、ケバブケースの文字列を返却します。
// 例： "CamelCaseString" -> "camel-case-string"
func ConvertCamelToKebab(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				sb.WriteRune('-')
			}
			sb.WriteRune(unicode.ToLower(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
