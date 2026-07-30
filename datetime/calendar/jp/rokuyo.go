package jp

import "time"

// rokuyoNames は六曜の名称を表す配列です。
// インデックスは（旧暦の月 + 旧暦の日）を 6 で割った剰余に対応します。
var rokuyoNames = [6]string{"大安", "赤口", "先勝", "友引", "先負", "仏滅"}

var lunisolar = NewJapaneseLunisolarCalendar()

// GregorianDateToRokuyoString はグレゴリオ暦の日付から六曜（大安・仏滅など）の名称を取得します。
// 六曜は旧暦の月と日から定まるため（旧暦 1 月 1 日は必ず先勝）、内部で旧暦変換を行います。
// 閏月は本来の月と同じ月番号として扱います（例: 閏6月は 6 月として計算）。
// 旧暦変換のサポート範囲（1960-01-28〜2050-01-22）外の日付にはエラーを返します。
func GregorianDateToRokuyoString(date time.Time) (string, error) {
	lunarYear, lunarMonth, lunarDay, err := lunisolar.gregorianToLunar(date)
	if err != nil {
		return "", err
	}

	leapMonth, err := lunisolar.GetLeapMonth(lunarYear)
	if err != nil {
		return "", err
	}

	// 閏月以降の通し番号を伝統的な月番号へ補正する
	if leapMonth > 0 && lunarMonth > leapMonth {
		lunarMonth--
	}

	return rokuyoNames[(lunarMonth+lunarDay)%6], nil
}
