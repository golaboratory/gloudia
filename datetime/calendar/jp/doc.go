// Package jp は日本の暦法に関する機能を提供します。
// 太陰太陽暦（旧暦）変換、元号判定、新暦ベースの和暦表記、六曜、二十四節気、
// 干支（年干支・日干支）、和風月名などをサポートします。
//
// 機能ごとの対応範囲:
//   - 旧暦変換・六曜 (JapaneseLunisolarCalendar, GregorianDateToRokuyoString):
//     1960-01-28〜2050-01-22（旧暦 1960〜2049 年）。検証済みの暦テーブルが
//     存在する範囲に限定しています。
//   - 新暦ベースの和暦 (GregorianDateToWareki, GregorianDateToWarekiString):
//     1873-01-01（明治6年、グレゴリオ暦採用日）以降。上限はありません。
//   - 二十四節気 (GregorianYearTo24SekkiList など): 近似式による計算で、
//     1900〜2100 年程度を想定。精度の詳細は sekki.go を参照してください。
//   - 干支・和風月名: 年月日の算術のみで計算でき、範囲制限はありません
//     （年は 0 以上）。
//
// 太陰太陽暦の変換ロジックは .NET (dotnet/runtime, MIT License) の
// JapaneseLunisolarCalendar 実装を参考にしています。詳細は lunisolar.go の
// ファイル冒頭の注記を参照してください。
//
// 本パッケージの日付判定はすべて time.Time の壁時計上の年月日で行われ、
// タイムゾーンによる瞬間の差は無視されます。JST の time.Time をそのまま渡せます。
package jp
