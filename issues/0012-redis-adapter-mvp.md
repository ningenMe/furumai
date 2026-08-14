---
status: done
created: 2026-08-14
related: docs/adapter-capability-catalog.md, issues/0004-http-adapter-mvp.md, issues/0009-mysql-adapter-mvp.md
---

# 0012: Redis (KVS) adapter実装

`docs/adapter-capability-catalog.md`のKVS（Redis）セクションに基づき、
MVPスコープ後の拡張としてRedis adapterを実装する。

## やること

- Stimulus adapter: `Set(key, value, opts: TTL)` / `Del(key...)` /
  `Expire(key, ttl)` / `HSet`/`LPush`/`SAdd`/`ZAdd` / `FlushDB`
  （危険な操作なので明示的opt-in必須）
- Observation: パターンにマッチする`map[string]Value`を取得し、
  `furumai.Diff`/`furumai.ThenEqual`でそのまま構造比較できるように
  する（`Value`はkeyのRedis型に応じて`string`/`map[string]string`/
  `[]string`/`map[string]float64`のいずれかを持つ`any`）

## 制約

- MySQL/Kafkaと異なり、driverライブラリの追加は行わない。RESP
  （Redisのwire protocol）は単純なテキストベースプロトコルであり、
  「標準ライブラリで実現できることは標準ライブラリで行う」という
  依存最小化方針の原則に照らして、`net`/`bufio`による最小限の
  自前実装が妥当と判断した（MySQL/Kafkaのように複雑なバイナリ
  プロトコルではないため、この判断はKVSに固有）
- サポートするコマンドはカタログに記載のものに限定する
  （Pub/Sub、Lua script、cluster対応等は対象外）
- key走査は`KEYS`コマンドを使う（本番運用でのブロッキングリスクは
  周知だが、テスト用途の小規模DBを対象とするため許容する）
- Environment Manager（Redisサーバーそのものの起動）は対象外
- ローカル環境（このrunner）にはRedis/Dockerが無いため、ローカルでの
  実サーバー接続確認は行わない。確認はCIのサービスコンテナに寄せる

## Definition of Done

- `Set`等でキー状態に刺激を与え、`Get`で取得した`map[string]Value`を
  `furumai.ThenEqual`で検証するサンプルテストがCI上で実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `adapter/redis`にRESPプロトコルの最小クライアント（`encodeCommand`/
  `readReply`）と、その上に`Stimulus`（`Set`/`Del`/`Expire`/`HSet`/
  `LPush`/`SAdd`/`ZAdd`/`FlushDB`/`Get`）を実装。外部依存ゼロ。
- `readReply`はRESPの5種類（simple string/error/integer/bulk
  string/array、nilのbulk string・array含む）をパースする。手書き
  プロトコル実装であるためバグの影響が大きく、`net.Conn`無しで
  `bufio.Reader`に直接canned RESPバイト列を読ませる形でunit test
  （`TestReadReply`等）を手厚くした。
- `Get`は`KEYS pattern`→各keyに`TYPE`→型ごとに`GET`/`HGETALL`/
  `LRANGE`/`SMEMBERS`/`ZRANGE ... WITHSCORES`を呼んで`Value`を
  構築している。TTL自体（`Within(min,max)`で検証したいケース）は
  `Value`に含めていない。将来`Get`の返り値を拡張する場合は要検討。
- `examples/redis_test.go`は`REDIS_ADDR`環境変数が無ければ`t.Skip`
  する統合テスト。CIに`redis:7-alpine`のサービスコンテナ
  （`redis-cli ping`によるhealthcheck付き）を追加した。
