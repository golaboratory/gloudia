package jp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGregorianDateToSekki_NoMatchReturnsEmpty(t *testing.T) {
	cases := []struct {
		name string
		date time.Time
	}{
		{"小寒と大寒の間", time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
		{"年初(節気でない)", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"月末(節気でない)", time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
		{"立春の翌日", time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := GregorianDateToSekki(c.date)
			require.NoError(t, err)
			assert.Equal(t, "", got, "節気に該当しない日は空文字列を返すべき")
		})
	}
}

func TestGregorianDateToSekki_IntraDayBoundary(t *testing.T) {
	// 小寒 = 2024-01-06。当日内の任意の時刻は該当し、翌日0時は非該当であること。
	midnight, err := GregorianDateToSekki(time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, "小寒", midnight)

	lateInDay, err := GregorianDateToSekki(time.Date(2024, 1, 6, 23, 59, 59, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, "小寒", lateInDay, "同日内の時刻はその節気に該当すべき")

	nextMidnight, err := GregorianDateToSekki(time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, "", nextMidnight, "翌日0時は半開区間の境界外で非該当であるべき")
}
