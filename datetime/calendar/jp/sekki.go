package jp

import (
	"time"
)

// sekki は二十四節気の計算に使用するパラメータを保持する構造体です。
type sekki struct {
	Month               int     // 節気が属する月
	DValue              float64 // 節気の日付計算用定数D
	AValue              float64 // 節気の日付計算用定数A
	YearAdjustmentValue int     // 年の補正値
}

// sekkiList は二十四節気の名称と対応するパラメータのマップです。
var sekkiList = map[string]sekki{
	"小寒": {Month: 1, DValue: 6.3811, AValue: 0.242778, YearAdjustmentValue: -1},
	"大寒": {Month: 1, DValue: 21.1046, AValue: 0.242765, YearAdjustmentValue: -1},
	"立春": {Month: 2, DValue: 4.8693, AValue: 0.242713, YearAdjustmentValue: -1},
	"雨水": {Month: 2, DValue: 19.7062, AValue: 0.242627, YearAdjustmentValue: -1},
	"啓蟄": {Month: 3, DValue: 6.3968, AValue: 0.242512, YearAdjustmentValue: 0},
	"春分": {Month: 3, DValue: 21.4471, AValue: 0.242377, YearAdjustmentValue: 0},
	"清明": {Month: 4, DValue: 5.6280, AValue: 0.242231, YearAdjustmentValue: 0},
	"穀雨": {Month: 4, DValue: 20.9375, AValue: 0.242083, YearAdjustmentValue: 0},
	"立夏": {Month: 5, DValue: 6.3771, AValue: 0.241945, YearAdjustmentValue: 0},
	"小満": {Month: 5, DValue: 21.9300, AValue: 0.241825, YearAdjustmentValue: 0},
	"芒種": {Month: 6, DValue: 6.5733, AValue: 0.241731, YearAdjustmentValue: 0},
	"夏至": {Month: 6, DValue: 22.2747, AValue: 0.241669, YearAdjustmentValue: 0},
	"小暑": {Month: 7, DValue: 8.0091, AValue: 0.241642, YearAdjustmentValue: 0},
	"大暑": {Month: 7, DValue: 23.7317, AValue: 0.241654, YearAdjustmentValue: 0},
	"立秋": {Month: 8, DValue: 8.4102, AValue: 0.241703, YearAdjustmentValue: 0},
	"処暑": {Month: 8, DValue: 24.0125, AValue: 0.241786, YearAdjustmentValue: 0},
	"白露": {Month: 9, DValue: 8.5186, AValue: 0.241898, YearAdjustmentValue: 0},
	"秋分": {Month: 9, DValue: 23.8896, AValue: 0.242032, YearAdjustmentValue: 0},
	"寒露": {Month: 10, DValue: 9.1414, AValue: 0.242179, YearAdjustmentValue: 0},
	"霜降": {Month: 10, DValue: 24.2487, AValue: 0.242328, YearAdjustmentValue: 0},
	"立冬": {Month: 11, DValue: 8.2396, AValue: 0.242469, YearAdjustmentValue: 0},
	"小雪": {Month: 11, DValue: 23.1189, AValue: 0.242592, YearAdjustmentValue: 0},
	"大雪": {Month: 12, DValue: 7.9152, AValue: 0.242689, YearAdjustmentValue: 0},
	"冬至": {Month: 12, DValue: 22.6587, AValue: 0.242752, YearAdjustmentValue: 0},
}

// GregorianYearTo24SekkiList は指定した西暦年の二十四節気の日付と名称のマップを返します。
// year: 西暦年
// 戻り値: 日付をキー、節気名を値とするマップ。エラーがあれば error を返します。
func GregorianYearTo24SekkiList(year int) (map[time.Time]string, error) {
	var result = make(map[time.Time]string)

	for k, v := range sekkiList {
		y := float64((year + v.YearAdjustmentValue) - 1900)
		dayOfMonth := int(v.DValue+(v.AValue*y)) - int(y/4)
		result[time.Date(year, time.Month(v.Month), dayOfMonth, 0, 0, 0, 0, time.UTC)] = k
	}

	return result, nil
}

// GregorianDateToSekki は指定した日付がどの節気に該当するかを返します。
// dt: 判定する日付
// 戻り値: 節気名。該当しない場合は空文字列。エラーがあれば error を返します。
func GregorianDateToSekki(dt time.Time) (string, error) {
	year := dt.Year()
	sekkiList, err := GregorianYearTo24SekkiList(year)
	if err != nil {
		return "", err
	}

	for date, sekki := range sekkiList {
		if dt.Equal(date) || (dt.After(date) && dt.Before(date.AddDate(0, 0, 1))) {
			return sekki, nil
		}
	}

	return "", nil
}
