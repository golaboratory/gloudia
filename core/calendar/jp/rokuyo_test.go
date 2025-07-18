package jp

import (
	"testing"
	"time"
)

func TestGregorianDateToRokuyoString(t *testing.T) {
	type testCase struct {
		date     time.Time
		expected string
		wantErr  bool
	}

	tests := []testCase{
		// 1960年1月28日（昭和35年旧暦1月1日）: 1+1=2, 2%6=2 → 先勝
		{time.Date(1960, 1, 28, 0, 0, 0, 0, time.UTC), "先勝", false},
		// 1960年2月1日（旧暦1月5日）: 1+5=6, 6%6=0 → 大安
		{time.Date(1960, 2, 1, 0, 0, 0, 0, time.UTC), "大安", false},
		// 1960年6月15日（旧暦5月20日）: 5+20=25, 25%6=1 → 赤口
		{time.Date(1960, 6, 15, 0, 0, 0, 0, time.UTC), "赤口", false},
		// 1960年7月18日（旧暦6月25日, 閏月前）: 6+25=31, 31%6=1 → 赤口
		{time.Date(1960, 7, 18, 0, 0, 0, 0, time.UTC), "赤口", false},
		// 1960年8月1日（旧暦閏6月8日, 閏月中）: 7+8=15, 15%6=3 → 友引
		{time.Date(1960, 8, 1, 0, 0, 0, 0, time.UTC), "友引", false},
		// 1960年8月30日（旧暦閏6月37日, 閏月末）: 7+37=44, 44%6=2 → 先勝
		{time.Date(1960, 8, 30, 0, 0, 0, 0, time.UTC), "先勝", false},
		// 1960年9月1日（旧暦7月1日, 閏月後）: 6+1=7, 7%6=1 → 赤口
		{time.Date(1960, 9, 1, 0, 0, 0, 0, time.UTC), "赤口", false},
		// 1989年1月8日（平成元年旧暦12月1日）: 12+1=13, 13%6=1 → 赤口
		{time.Date(1989, 1, 8, 0, 0, 0, 0, time.UTC), "赤口", false},
		// 1989年2月6日（平成元年旧暦1月1日）: 1+1=2, 2%6=2 → 先勝
		{time.Date(1989, 2, 6, 0, 0, 0, 0, time.UTC), "先勝", false},
		// 1995年8月28日（旧暦閏8月4日, 閏月中）: 9+4=13, 13%6=1 → 赤口
		{time.Date(1995, 8, 28, 0, 0, 0, 0, time.UTC), "赤口", false},
		// 2006年7月25日（旧暦閏7月1日, 閏月中）: 8+1=9, 9%6=3 → 友引
		{time.Date(2006, 7, 25, 0, 0, 0, 0, time.UTC), "友引", false},
		// 2014年9月24日（旧暦閏9月1日, 閏月中）: 10+1=11, 11%6=5 → 仏滅
		{time.Date(2014, 9, 24, 0, 0, 0, 0, time.UTC), "仏滅", false},
		// 2019年5月1日（令和元年旧暦3月27日）: 3+27=30, 30%6=0 → 大安
		{time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC), "大安", false},
		// 2023年3月22日（旧暦閏2月1日, 閏月中）: 3+1=4, 4%6=4 → 先負
		{time.Date(2023, 3, 22, 0, 0, 0, 0, time.UTC), "先負", false},
		// 2023年4月19日（旧暦閏2月29日, 閏月中）: 3+29=32, 32%6=2 → 先勝
		{time.Date(2023, 4, 19, 0, 0, 0, 0, time.UTC), "先勝", false},
		// 2023年4月20日（旧暦3月1日, 閏月後）: 2+1=3, 3%6=3 → 友引
		{time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC), "友引", false},
		// 2044年8月1日（旧暦閏7月1日, 閏月中）: 8+1=9, 9%6=3 → 友引
		{time.Date(2044, 8, 1, 0, 0, 0, 0, time.UTC), "友引", false},
		// 2049年12月31日（旧暦12月7日）: 12+7=19, 19%6=1 → 赤口
		{time.Date(2049, 12, 31, 0, 0, 0, 0, time.UTC), "赤口", false},
		// サポート外（1959年12月31日）はエラー
		{time.Date(1959, 12, 31, 0, 0, 0, 0, time.UTC), "", true},
		// サポート外（2050年1月23日）はエラー
		{time.Date(2050, 1, 23, 0, 0, 0, 0, time.UTC), "", true},
		// サポート外（3000年1月1日）はエラー
		{time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), "", true},
		// 1961年2月15日（旧暦2月1日）: 2+1=3, 3%6=3 → 友引
		{time.Date(1961, 2, 15, 0, 0, 0, 0, time.UTC), "友引", false},
		// 1976年9月1日（旧暦8月6日, 閏月前）: 8+6=14, 14%6=2 → 先勝
		{time.Date(1976, 9, 1, 0, 0, 0, 0, time.UTC), "先勝", false},
		// 1984年11月23日（旧暦10月1日, 閏月後）: 10+1=11, 11%6=5 → 仏滅
		{time.Date(1984, 11, 23, 0, 0, 0, 0, time.UTC), "仏滅", false},
		// 2011年2月3日（旧暦1月1日）: 1+1=2, 2%6=2 → 先勝
		{time.Date(2011, 2, 3, 0, 0, 0, 0, time.UTC), "先勝", false},
	}

	for i, tt := range tests {
		got, err := GregorianDateToRokuyoString(tt.date)
		if (err != nil) != tt.wantErr {
			t.Errorf("case %d: GregorianDateToRokuyoString(%v) error = %v, wantErr %v", i, tt.date, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.expected {
			t.Errorf("case %d: GregorianDateToRokuyoString(%v) = %v, want %v", i, tt.date, got, tt.expected)
		}
	}
}
