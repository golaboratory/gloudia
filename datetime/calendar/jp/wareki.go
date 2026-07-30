package jp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/newmo-oss/ergo"
)

// ErrDateBeforeWareki は新暦ベースの和暦変換のサポート範囲
// （1873-01-01 = 明治6年1月1日、日本のグレゴリオ暦採用日）より前の
// 日付が指定された場合のエラーです。
var ErrDateBeforeWareki = ergo.NewSentinel("date is before the supported wareki range")

// warekiMinDate は新暦ベースの和暦変換のサポート下限日です。
// 1872 年以前の日本の公式暦は太陰太陽暦（天保暦）であり、グレゴリオ暦の
// 年月日をそのまま和暦に対応付けられないため、改暦日以降に限定します。
var warekiMinDate = time.Date(1873, 1, 1, 0, 0, 0, 0, time.UTC)

// warekiEras は新暦ベースの和暦変換に使用する元号定義です（新しい順）。
// 元号IDは .NET の JapaneseCalendar と同じ番号体系です。
var warekiEras = []EraInfo{
	{5, "令和", "Reiwa", time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), 2018},
	{4, "平成", "Heisei", time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), 1988},
	{3, "昭和", "Showa", time.Date(1926, 12, 25, 0, 0, 0, 0, time.UTC), 1925},
	{2, "大正", "Taisho", time.Date(1912, 7, 30, 0, 0, 0, 0, time.UTC), 1911},
	{1, "明治", "Meiji", time.Date(1868, 10, 23, 0, 0, 0, 0, time.UTC), 1867},
}

// GregorianDateToWareki は指定した日付（壁時計上の年月日）に対応する
// 新暦ベースの和暦（元号情報と元号年）を返します。
// 旧暦ベースの JapaneseLunisolarCalendar.GetYear と異なり、年の区切りは
// グレゴリオ暦の 1 月 1 日で、元号の切り替わりは改元日当日です
// （例: 1989-01-07 は昭和64年、1989-01-08 は平成元年）。
// 1873-01-01（明治6年、グレゴリオ暦採用日）より前の日付には
// ErrDateBeforeWareki をラップしたエラーを返します。
func GregorianDateToWareki(t time.Time) (EraInfo, int, error) {
	d := toUTCDate(t)
	if d.Before(warekiMinDate) {
		return EraInfo{}, 0, ergo.Wrap(ErrDateBeforeWareki, d.Format("2006-01-02"))
	}
	for _, era := range warekiEras {
		if !d.Before(era.StartDate) {
			return era, d.Year() - era.YearOffset, nil
		}
	}
	// warekiMinDate 以降の日付は必ず明治以降のいずれかの元号に該当するため、
	// ここには到達しない（防御的リターン）。
	return EraInfo{}, 0, ergo.Wrap(ErrDateBeforeWareki, d.Format("2006-01-02"))
}

// GregorianDateToWarekiString は指定した日付を「令和8年7月30日」形式の
// 和暦文字列に変換します。元号 1 年目は慣用に従い「元年」と表記します
// （例: 2019-05-01 は「令和元年5月1日」）。
// サポート範囲は GregorianDateToWareki と同じです。
func GregorianDateToWarekiString(t time.Time) (string, error) {
	era, year, err := GregorianDateToWareki(t)
	if err != nil {
		return "", err
	}
	yearStr := strconv.Itoa(year)
	if year == 1 {
		yearStr = "元"
	}
	d := toUTCDate(t)
	return fmt.Sprintf("%s%s年%d月%d日", era.Name, yearStr, int(d.Month()), d.Day()), nil
}
