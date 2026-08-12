# Furumai (振る舞い)

Furumaiは、サーバーサイドシステムの振る舞いを、実装言語やフレームワークに依存せずテストするためのフレームワーク。

> Stimulate the system, observe its behavior, and verify the result.

HTTP API、DB、Kafka、shell commandなど、サーバーサイドで発生する処理を「システムへの刺激」として扱い、その結果として観測される状態を検証する。

## Setup

```sh
go get github.com/ningenMe/furumai
```

## Usage

テストは `given` / `when` / `then` の3ステップで書く。

- `given`: テストの前提条件を整える（DB seed、事前のKafka publishなど）
- `when`: 検証対象そのものに刺激を与える（HTTP request、shell commandなど）
- `then`: 刺激の結果を観測し、期待するフルステートと構造比較する

```go
package examples

import (
	"net/http"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/rest"
)

func TestGreetingAPI(t *testing.T) {
	client := rest.NewStimulus("http://localhost:8080")

	var resp *rest.Response

	furumai.When(t, func() error {
		var err error
		resp, err = client.Get("/greeting", rest.WithQuery("name", "Alice"))
		return err
	})

	furumai.ThenEqual(t, *resp, rest.Response{
		StatusCode: http.StatusOK,
		Headers:    furumai.Ignore(),
		Body:       `{"greeting":"hello, Alice"}`,
	})
}
```

`given`と`when`は同じStimulus adapter（上の例では`rest.Stimulus`）を共用する。プロトコルごとのadapterは`adapter/`配下のsubpackageに分かれている(HTTPは`adapter/rest`)。`then`は期待する完全な状態を1つの構造体として書き、`furumai.ThenEqual`が実際の状態との差分を全てまとめて報告する。値の一部だけを確認したい場合は`furumai.Any()`/`Regex()`/`Within()`/`Ignore()`/`AnyOrder()`といったmatcherを埋め込める。

テストの実行は通常の`go test`。Parameterized testもGoのtable-driven testパターンでそのまま書ける。より多くのサンプルは[`examples/`](./examples)を参照。

### furumaiコマンド

```sh
go install github.com/ningenMe/furumai/cmd/furumai@latest

furumai version   # バージョン表示
furumai help      # ヘルプ表示
```

## License

[MIT](./LICENSE)
