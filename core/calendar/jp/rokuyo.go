package jp

import "time"

// rokuyoNames は六曜のインデックスと名称の対応を表します。
var rokuyoNames = map[int]string{
	0: "大安",
	1: "赤口",
	2: "先勝",
	3: "友引",
	4: "先負",
	5: "仏滅",
}

// GregorianDateToRokuyoString はグレゴリオ暦の日付から六曜（大安・仏滅など）の名称を取得します。
func GregorianDateToRokuyoString(date time.Time) (string, error) {
	lunisolar := NewJapaneseLunisolarCalendar()

	//lunisolarYear, err := lunisolar.GetYear(date)
	//if err != nil {
	//		return "", err
	//}

	lunisolarMonth, err := lunisolar.GetMonth(date)
	if err != nil {
		return "", err
	}

	lunisolarDay, err := lunisolar.GetDayOfMonth(date)
	if err != nil {
		return "", err
	}

	leapMonth, err := lunisolar.GetLeapMonth(date.Year())
	if err != nil {
		return "", err
	}

	if leapMonth > 0 && lunisolarMonth > leapMonth {
		lunisolarMonth--
	}

	rokuyoIndex := (lunisolarMonth + lunisolarDay) % 6
	return rokuyoNames[rokuyoIndex], nil
}
