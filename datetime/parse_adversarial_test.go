package datetime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlexibleDate_ReturnsUTCLocation(t *testing.T) {
	for _, in := range []string{"2026-04-25", "2026年4月25日"} {
		in := in
		t.Run(in, func(t *testing.T) {
			got, err := ParseFlexibleDate(in)
			require.NoError(t, err)
			assert.Equal(t, "UTC", got.Location().String(), "ゾーン無しレイアウトは UTC を返すべき")
			_, offset := got.Zone()
			assert.Equal(t, 0, offset, "UTC のオフセットは 0 であるべき")
		})
	}
}
