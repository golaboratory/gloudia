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

// ConvertToEraDate は与えられた日付を和暦の日付文字列に変換します。
// 引数:
//   - dt: 変換対象の日付
//
// 戻り値:
//   - string: 和暦の日付文字列
//   - error: エラー情報
func ConvertToEraDate(dt time.Time) (string, error) {
	return ConvertToEraDateWithFormat(dt, "ggggyy年MM月dd日")
}

// ConvertToEraDateWithFormat は与えられた日付を指定されたフォーマットの和暦日付文字列に変換します。
// 引数:
//   - dt: 変換対象の日付
//   - format: フォーマット文字列
//
// 戻り値:
//   - string: 和暦の日付文字列
//   - error: エラー情報
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

// ConvertToZodiacByYear は与えられた年を干支に変換します。
// 引数:
//   - year: 干支に変換する年
//
// 戻り値:
//   - string: 干支の文字列
//   - error: エラー情報
func ConvertToZodiacByYear(year int) (string, error) {
	kan, shi := zodiac.ZodiacYearNumber(year)
	return fmt.Sprintf("%v%v", kan, shi), nil
}

// ConvertToZodiacByDate は与えられた日付の年を干支に変換します。
// 引数:
//   - dt: 干支に変換する日付
//
// 戻り値:
//   - string: 干支の文字列
//   - error: エラー情報
func ConvertToZodiacByDate(dt time.Time) (string, error) {
	return ConvertToZodiacByYear(dt.Year())
}

// ConvertToDayZodiacByDate は与えられた日付を日干支に変換します。
// 引数:
//   - dt: 日干支に変換する日付
//
// 戻り値:
//   - string: 日干支の文字列
//   - error: エラー情報
func ConvertToDayZodiacByDate(dt time.Time) (string, error) {
	t := value.NewDate(dt)
	kan, shi := zodiac.ZodiacDayNumber(t)
	return fmt.Sprintf("%v%v", kan, shi), nil
}

// ConvertCamelToKebab はキャメルケースの文字列を受け取り、ケバブケースの文字列を返却します。
// 例: "CamelCaseString" -> "camel-case-string"
// 引数:
//   - s: キャメルケースの文字列
//
// 戻り値:
//   - string: ケバブケースの文字列
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
