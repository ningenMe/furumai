# Furumai (振る舞い)

Furumaiは、サーバーサイドシステムの振る舞いを、実装言語やフレームワークに依存せずテストするためのフレームワーク。

> Stimulate the system, observe its behavior, and verify the result.

HTTP API、scheduled batch、Kafka subscriber、queue worker、CLI、webhookなど、サーバーサイドで発生する処理を「システムへの刺激」として扱い、その結果として観測される状態やイベントを検証する。設計方針の詳細は [`docs/core-design-direction.md`](./docs/core-design-direction.md)、経緯は [GitHub Issue #1](https://github.com/ningenMe/furumai/issues/1) を参照。

## DSL

テストは `given` / `when` / `then` の3ステップで構成する。

- `given`: テストの前提条件を整えるための刺激を与える（DB seed、事前のKafka publishなど）
- `when`: 検証対象そのものの刺激を与える（HTTP request、Kafka publish、shell commandなど）
- `then`: 刺激の結果を観測・検証する（HTTP response、DB state、stdoutなど）

`given`/`when`/`then` は特定のプロトコルに縛られない汎用的なステップとして設計している。`given`と`when`はどちらも同じ種類の操作（Stimulus）で、違いはシナリオ内での役割だけ。現時点ではHTTP/DB/Kafkaといった具体的なadapterはまだ無く、各ステップの中身は呼び出し側が直接書く関数になる（adapterは今後のタスクで追加していく）。

```go
package examples

import (
	"fmt"
	"testing"

	"github.com/ningenMe/furumai"
)

func TestExample(t *testing.T) {
	var got string

	furumai.Given(t, func() error {
		got = ""
		return nil
	})

	furumai.When(t, func() error {
		got = "hello"
		return nil
	})

	furumai.Then(t, func() error {
		if got != "hello" {
			return fmt.Errorf("got %q", got)
		}
		return nil
	})
}
```

`given`/`when`/`then` は `*testing.T` を受け取る薄い関数で、実行自体はGo標準の `go test` にそのまま乗せている（独自の実行エンジンはまだ無い）。Parameterized testはGoのtable-driven testパターンで書ける。サンプル一式は [`examples/`](./examples) を参照。

## Usage

```sh
go build ./...
./furumai version   # バージョン表示
./furumai help      # ヘルプ表示

go test ./examples/...   # テスト実行(今はこちらが実体)
```

`furumai test` のような専用のテスト実行コマンドはまだ無い。

## License

[MIT](./LICENSE)
