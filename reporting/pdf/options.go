package pdf

// ConvertOptions はPDF変換時のオプション設定です
type ConvertOptions struct {
	// Landscape をtrueにすると横向きで出力します
	Landscape bool

	// PageRanges は変換するページ範囲を指定します (例: "1-5", "1,3,5")
	// 空文字の場合は全ページ変換されます
	PageRanges string

	// Scale は拡大縮小率を指定します (例: 1.0 が標準)
	// 0の場合はデフォルト(指定なし)となります
	Scale float64
}

// DefaultOptions は標準的な設定を返します
func DefaultOptions() *ConvertOptions {
	return &ConvertOptions{
		Landscape:  false,
		PageRanges: "",
		Scale:      0,
	}
}
