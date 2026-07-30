// Package datetime は日付・時刻に関する汎用ユーティリティを提供します。
//
// 提供機能:
//   - ParseFlexibleDate: ISO8601 (`2006-01-02`)、日本語形式 (`2006年1月2日`)、
//     スラッシュ区切り (`2006/1/2`)、8 桁数字 (`20060102`) のフォールバックパース
//   - 日付フォーマット定数 (ISO8601DateFormat / JPDateFormat / SlashDateFormat /
//     CompactDateFormat)
//   - JST: 日本標準時 (UTC+9) の固定タイムゾーン
//
// 注意: 日本語日付形式のパースを含むため、本パッケージはロケール非依存ではありません。
// 和暦・暦注（元号・旧暦・六曜・干支・節気等）の特化機能は `datetime/calendar/jp` を
// 参照してください。
package datetime
