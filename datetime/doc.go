// Package datetime は日付・時刻に関する汎用ユーティリティを提供します。
//
// 提供機能:
//   - ParseFlexibleDate: ISO8601 (`2006-01-02`) と日本語形式 (`2006年1月2日`) の
//     フォールバックパース
//   - 日付フォーマット定数 (ISO8601DateFormat / JPDateFormat)
//
// 注意: 日本語日付形式のパースを含むため、本パッケージはロケール非依存ではありません。
// 和暦（六曜・干支・節気等）の特化機能は `datetime/calendar/jp` を参照してください。
package datetime
