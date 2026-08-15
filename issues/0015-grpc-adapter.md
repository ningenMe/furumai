---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md
---

# 0015: gRPC adapter実装

`docs/adapter-capability-catalog.md`のRPCセクション（gRPC部分）に
基づき、gRPC adapterを実装する。

## やること

- Stimulus adapter: Unary callを実行できるようにする
- Observation: `Response{Message, StatusCode, Trailer}`というフル
  ステートを1つ返す。`furumai.Diff`/`furumai.ThenEqual`でそのまま
  構造比較できるようにする（全フィールドmatcher埋め込みが効くよう
  `any`型で持つ）

## 制約

- Streaming call（catalogに記載あり）は今回のスコープ外。Unaryのみ
- gRPCのwire protocol（HTTP/2 + protobuf framing）を自前実装するのは
  非現実的なため、`google.golang.org/grpc` + `google.golang.org/protobuf`
  を明示的な例外として許容する（DB driver・Kafka clientと同じ扱い）
- adapterは呼び出し側の生成済み`.proto`スタブ（`proto.Message`実装）を
  そのまま受け取る形にし、動的解決（server reflection等）は行わない

## Definition of Done

- Unary RPCを呼び出し、`Response{Message, StatusCode, Trailer}`を
  `furumai.ThenEqual`で検証するサンプルテストが実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `adapter/grpc`に`Stimulus`（`Unary`）と`Response{Message,
  StatusCode, Trailer}`を実装。非OKステータスは`Response.StatusCode`
  で表現し、Goの`error`にはしない（RPCとして評価できなかった場合
  ─接続失敗等─のみ`error`を返す）。
- 依存は`google.golang.org/grpc`+`google.golang.org/protobuf`、
  `go mod tidy`後のindirectは`x/net`/`x/sys`/`x/text`/
  `genproto/googleapis/rpc`の4つ。他adapterより重いが、gRPCの
  wire protocolを自前実装するのは非現実的なため、DB driver/Kafka
  clientと同格の必須例外として扱った。
- unit test（`adapter/grpc/grpc_test.go`）と`examples/grpc_test.go`は
  どちらも`.proto`/codegen無しで完結させた。`grpc.ServiceDesc`を
  手書きし、`google.golang.org/protobuf/types/known/wrapperspb`
  （protobuf runtime同梱の生成済み型）をrequest/reply代わりに使う
  ことで、外部サービス・生成コード無しに実gRPCサーバー/クライアント
  往復を検証できている。実際の利用者は自分の生成済みスタブを渡す
  想定。
- パッケージ名は`grpc`（`google.golang.org/grpc`と同名）。両方を
  同じファイルでimportする場合は片方をaliasする必要がある旨を
  package docに明記した（`adapter/rest`のように別名にできる自然な
  代替語が無いため、この1件は衝突を許容する判断にした）。
