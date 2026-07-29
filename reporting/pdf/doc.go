// Package pdf は Gotenberg API を使用した Excel → PDF 変換機能を提供します。
// io.Pipe によるストリーミング処理でメモリ効率の高い大容量ファイル変換を実現します。
//
// # Gotenberg セットアップ手順
//
// 1. Docker で Gotenberg を起動:
//
//	docker run --rm -p 3000:3000 gotenberg/gotenberg:8
//
// 2. API エンドポイント: http://localhost:3000
//
// 3. ヘルスチェック:
//
//	curl http://localhost:3000/health
//
// 4. 使用例:
//
//	client := pdf.NewClient("http://localhost:3000")
//	pdfBytes, err := client.ConvertExcelToPDF(ctx, excelBytes, "report.xlsx")
//
// 本番環境では Docker Compose または Kubernetes でサイドカーとしてデプロイすることを推奨します。
// compose.yml での設定例:
//
//	services:
//	  gotenberg:
//	    image: gotenberg/gotenberg:8
//	    restart: unless-stopped
//	    ports:
//	      - "3000:3000"
package pdf
