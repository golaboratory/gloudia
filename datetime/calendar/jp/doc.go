// Package jp は日本の暦法に関する機能を提供します。
// 太陰太陽暦（旧暦）変換、元号判定、六曜、二十四節気、干支（年干支・日干支）、
// 月齢名称などをサポートします。対応範囲は 1960年〜2100年 です。
//
// 太陰太陽暦の変換ロジックは .NET (dotnet/runtime, MIT License) の
// JapaneseLunisolarCalendar 実装を参考にしています。詳細は lunisolar.go の
// ファイル冒頭の注記を参照してください。
//
// 二十四節気は近似式による計算です。精度の詳細は sekki.go を参照してください。
package jp
