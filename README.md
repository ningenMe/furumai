# Furumai (振る舞い)

Furumaiは、サーバーサイドシステムの振る舞いを、実装言語やフレームワークに依存せずテストするためのフレームワーク。

コンセプトや設計方針は [GitHub Issue #1](https://github.com/ningenMe/furumai/issues/1) を参照。

## ステータス

v0開発中(private)。タスク管理は GitHub Issues ではなく [`issues/`](./issues/README.md) ディレクトリで行っている。
設計方針の詳細は [`docs/core-design-direction.md`](./docs/core-design-direction.md) を参照。

## 現在のUsage

まだMVPの土台部分のみで、機能は最小限。

### CLI

```sh
go build ./...
./furumai version   # バージョン表示
./furumai help      # ヘルプ表示
```

`furumai test` のような専用のテスト実行コマンドはまだ無い（後述の通り、
現状は `go test` にフリーライドしている）。

### テストの書き方（DSL）

`given`/`when`/`then` は `*testing.T` を受け取る薄い関数として提供されて
おり、実行自体はGo標準の `go test` を使う（独自の実行エンジンはまだ
無い）。

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

```sh
go test ./examples/...
```

サンプル一式は [`examples/`](./examples) を参照。HTTP/DB/Kafkaなどの
adapterはまだ無く、`when`/`then` の中身は自前の関数呼び出しになる。

## License

[MIT](./LICENSE)
