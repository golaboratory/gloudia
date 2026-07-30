// Package worker はバックグラウンドジョブのポーリング・実行フレームワークを提供します。
// Worker インターフェースでジョブキューとの通信を抽象化し、
// Processor でジョブタイプごとの処理ロジックをルーティングします。
//
// NewWorker でワーカープロセスを生成し、WorkerProcess.Start でポーリングを開始します。
// ジョブタイプごとの処理は JobProcessor のマップとして NewWorker（内部で NewProcessor）に登録します。
package worker
