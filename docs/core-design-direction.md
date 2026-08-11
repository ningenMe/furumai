---
status: draft
created: 2026-08-11
related: https://github.com/ningenMe/furumai/issues/1, issues/0001-core-design-direction.md
---

# Furumai — Core Design Direction

[GitHub Issue #1](https://github.com/ningenMe/furumai/issues/1)「方針」を前提に、実装着手前の設計整理を行う。

本ドキュメントの目的はコード量を増やすことではなく、**Furumaiの思想
（Stimulate the system, observe its behavior, and verify the result.）
を実現するための最小で強い設計を見つけること**である。

未決事項（テスト記述言語、DSL構文、内部アーキテクチャ、adapter設計、container
runtime、plugin architecture、application lifecycle management、assertion
API）は本ドキュメントでも確定しない。複数案を比較し、Furumaiの価値を最も強く
実現する案を推奨として示すに留める。

---

## 1. Core concept

Furumaiのcore conceptは、issue #1が掲げる以下の一文に集約される。

> **Stimulate the system, observe its behavior, and verify the result.**

これをフレームワークの設計原則として分解すると、次の3点になる。

1. **刺激（stimulus）とプロトコルの分離**
   `when` はHTTP requestに限定されない。shell command、DB操作、Kafka
   publish、queue投入など、「システムに変化を起こす操作」全般を同じ抽象
   （Stimulus）として扱う。
2. **観測（observation）とプロトコルの分離**
   `then` も同様に、HTTP response、DB state、Kafka message、stdout/exit
   codeなど「外部から観測可能な状態」全般を同じ抽象（Observation）として
   扱う。
3. **Black-boxであること**
   internal class / ORM model / framework-specific APIに依存しない。
   Furumaiが知っているのは「何を起こせるか（Stimulus adapter）」と「何を
   観測できるか（Observation adapter）」だけであり、対象システムの内部
   実装には一切関与しない。

この結果、Furumaiは「API testing framework」でも「既存BDD/E2E framework
の模倣」でもなく、**任意のプロトコルのStimulus/Observationをプラガブルな
adapterとして扱う、汎用的な"システム刺激応答検証エンジン"** として位置づけ
られる。given/when/thenという構造自体は目新しくないが、`when`/`then`が
特定プロトコルに縛られない点、および高速な並列実行がコアバリューである点が、
既存フレームワークとの差別化要因になる。

## 2. 想定されるユーザー体験（UX）

典型的な利用フローを想定する。

1. **導入**: Furumai CLIをインストールする（対象言語・フレームワークとは
   独立した単一バイナリ、または対応言語のパッケージマネージャ経由）。
2. **環境定義**: プロジェクト設定に、テストに必要な周辺基盤
   （PostgreSQL、Kafkaなど）を宣言する。アプリケーション本体は対象外
   （既に起動している前提）。
3. **テスト記述**: 型付き言語（またはそのDSL）で `given / when / then`
   ブロックを記述する。IDEの補完・型チェックが効き、YAML/JSONの設定ファイル
   を書く感覚ではなく、プログラムを書く感覚に近い。
4. **Parameterized test**: 同一シナリオを複数の入力パターンで検証する
   テストを、コピペではなくデータ駆動で自然に書ける。
5. **実行**: `furumai test` のようなコマンドで実行する。Furumaiが
   必要な周辺基盤をコンテナとして自動起動し、依存関係のないテストは並列に、
   依存関係が明示されたテストは直列に実行する。
6. **結果確認**: 失敗したテストについて、`when` で何を行い、`then` の
   どのassertionが期待値とどう異なったかが一目でわかるレポートが出力される。
   CI向けの構造化フォーマット（JUnit XML等）も出力できる。
7. **反復**: 単一テストの再実行、変更検知による再実行など、開発中の
   フィードバックループが速い。

利用者が重視するのは「大量のシステムテストを、周辺基盤のセットアップに
悩まされず、短時間で繰り返し実行できること」であり、UXの核心は
**「セットアップの手間の少なさ」と「実行の速さ」と「テストコードの
書き味の良さ」** の3点に集約される。

## 3. テスト記述方式の候補と比較

要件（静的型付け、IDE補完、型チェック、読みやすさ、parameterized test、
再利用性）を満たす記述方式を比較する。

| 候補 | 概要 | 静的型付け | IDE補完 | Parameterized test | 実装コスト | 備考 |
|---|---|---|---|---|---|---|
| A. 独自言語（外部DSL、専用パーサ/コンパイラ） | Furumai専用の新しいプログラミング言語を作る | ◎（設計次第） | △（LSP自作が必要） | ◎（言語機能として設計可） | 非常に高い（parser, type checker, LSP, formatter…を継続保守） | 表現力は最大だが、エコシステム構築コストが莫大 |
| B. 宣言的設定ファイル（YAML/JSON + schema） | テストをデータとして記述 | △（schemaによる疑似型付け止まり） | △（schema対応editorのみ） | ○（配列で表現可） | 低い | issueで明示的に「単なる設定ファイルにしない」と否定されている |
| C. Gherkin等の既存BDD形式 + step定義 | 自然言語シナリオ＋別言語のstep実装 | ×（シナリオ本体は非型付け文字列） | × | △（Scenario Outlineはあるが弱い） | 中 | 既存BDD frameworkの模倣であり、issueの制約に反する |
| D. 内部DSL（既存の静的型付け言語のtrailing lambda/builder構文を利用） | 既存言語のホスト機能（trailing lambda、fluent API等）でgiven/when/thenを表現 | ◎（ホスト言語の型システムをそのまま利用） | ◎（既存LSP/IDEをそのまま利用） | ◎（ホスト言語のループ・ジェネリクス・データクラスで自然に表現） | 中（パーサ自作不要、ライブラリ実装のみ） | 独自言語のコストを払わずに要件をほぼ満たせる |

**比較の要点**

- B・Cはissueの制約（「単なる設定ファイルにしない」「既存BDDの模倣を
  避ける」）に抵触するため候補から除外する。
- Aは表現力・独自性は最大だが、LSP・formatter・パッケージ管理まで
  自前で作り続ける必要があり、「MVPで最小に検証する」という目的に反する
  コストを伴う。
- Dは、静的型付け・IDE補完・型チェックという要件を、既存言語のツール
  チェーンにフリーライドすることで低コストに満たせる。given/when/then
  の入れ子構造は、trailing lambda構文を持つ言語（例: Kotlin）や
  アロー関数を持つ言語（例: TypeScript）で自然に表現できる。実際issue
  本文のサンプルコード自体が、trailing lambda DSL（Kotlin的な構文）に
  近い見た目をしている。

**推奨**: **D（内部DSL）**。独自言語（A）のコストとリスクを避けつつ、
issueが求める「高級言語に近い書き味」を実現できる。ホスト言語の具体候補
は4章で比較する。

なお、DSLの具体的なsyntax・ホスト言語そのものはここでは確定しない。
4章の技術スタック選定と合わせて決定する。

## 4. 技術スタックの候補と比較

Furumai自身の実装言語（エンジン、CLI、環境管理）の候補を比較する。
評価軸は、並列実行のしやすさ（並列実行はコアバリュー）、配布のしやすさ
（単一バイナリ等）、周辺エコシステム成熟度（DB/Kafka client、
container/testcontainers系ライブラリ）、およびテスト記述DSLのホスト
言語としての適性（3章）である。

| 候補 | 並行処理モデル | 配布 | Testcontainers系ライブラリ成熟度 | DSLホストとしての適性 | 備考 |
|---|---|---|---|---|---|
| Go | goroutine（軽量・実績豊富） | 単一静的バイナリ | 高い（testcontainers-go） | 低い（trailing lambda構文がなく、given/when/thenのネスト表現がやや冗長） | CLIツールとしての配布は最強クラス |
| Rust | async/tokio（高性能） | 単一静的バイナリ | 発展途上（testcontainers-rs） | 中（クロージャ構文はあるが所有権/lifetimeの学習コストがDSL利用者にも波及） | 性能は最高だが開発速度・コントリビューションコストが高い |
| Kotlin (JVM) | coroutine（構造化並行性） | JAR（+ GraalVM native-imageで単一バイナリ化可） | 最高（Testcontainers発祥のJVMエコシステム） | 高い（trailing lambda構文がgiven/when/thenと自然に一致、null安全、data class） | DSLホストとエンジン実装言語を統一しやすい |
| TypeScript (Node.js) | event loop + async/await（I/O待ち中心の並列実行と親和性が高い） | npm、または単一バイナリ化ツール（pkg等） | 中程度（testcontainers-node） | 高い（アロー関数でネスト表現可、エコシステム最大級） | 認知度・採用ハードルの低さが強み |

**比較の要点**

- Furumaiのテストは「対象システムへの刺激とその観測」がI/O待ち中心
  であり、CPUバウンドな並列処理性能そのもの（Go/Rustの得意領域）が
  決定的な差別化要因にはなりにくい。むしろ、3章で推奨した「内部DSL」の
  ホスト言語としての書き味と、環境管理に必須のcontainer/testcontainers
  エコシステムの成熟度の方が価値に直結する。
- issue本文のgiven/when/thenサンプルは、trailing lambda構文を持つ
  言語で最も自然に書ける。KotlinはTestcontainers発祥のJVMエコシステム
  でもあり、DB/Kafka clientの成熟度も高い。
- Go/Rustは環境管理・実行エンジン単体としては強いが、DSLホストとしての
  書き味で劣り、かつMVPでは「DSLホスト言語」と「エンジン実装言語」を
  分離する複雑さ（プロセス間通信、plugin protocol等）を負うほどの
  メリットがまだない（10章の将来拡張で検討）。

**推奨**: MVPでは**Kotlin (JVM)** を、DSLホスト言語兼エンジン実装言語
として単一採用する（5章のOption Aに対応）。配布上のJVM依存は、
GraalVM native-imageによる単一バイナリ化で緩和できる。TypeScriptは
採用ハードルの低さで次点の有力候補であり、将来的な多言語対応
（10章）の最初の追加候補としても位置づける。

## 5. Core architecture

DSLホスト言語とエンジン実装言語をどう分離するかで、2つのアーキテクチャ
案がある。

- **Option A: 単一言語モノリス** — DSL・実行エンジン・adapter・
  環境管理をすべて同一言語（4章の推奨: Kotlin）で実装する。プロセス
  間通信が不要で、型がプロセス境界をまたがず共有される。実装コストが
  最小で、MVPに向く。
- **Option B: エンジンとDSLの分離** — 実行エンジン・環境管理を
  systems言語（Go/Rust）、DSLホストを別言語（TypeScript等）とし、
  IPC（gRPC等）で接続する。真の意味での「テスト記述言語の多言語対応」
  を早期に実現できるが、plugin protocolの設計・維持コストをMVPの
  時点から負うことになる。

**推奨**: MVPは**Option A（単一言語モノリス）**。Option Bが提供する
「複数言語でテストを書ける」という価値は、v0のスコープ（1言語で
value propositionを検証する）には過剰であり、10章の将来拡張点として
持ち越す。

Option Aを前提としたレイヤー構成:

```
┌─────────────────────────────────────────────┐
│ Test Definition Layer（DSL）                  │
│  given/when/then ブロック、parameterized test、│
│  テストコードの合成・再利用                     │
├─────────────────────────────────────────────┤
│ Execution Engine                              │
│  テストグラフの構築、スケジューラ（並列/直列）    │
├───────────────────┬───────────────────────────┤
│ Stimulus Adapters   │ Observation Adapters      │
│ (when)              │ (then)                    │
│  - HTTP request      │  - HTTP response          │
│  - shell command      │  - DB state query         │
│  - DB操作             │  - Kafka message           │
│  - Kafka publish       │  - stdout/exit code        │
│  - queue投入            │  - Assertion API           │
├───────────────────┴───────────────────────────┤
│ Environment Manager                            │
│  container runtimeを介した周辺基盤の起動/停止/   │
│  ヘルスチェック（対象アプリ自体は管理しない）      │
├─────────────────────────────────────────────┤
│ Reporter                                       │
│  console出力、CI向け構造化レポート                │
└─────────────────────────────────────────────┘
```

Stimulus/Observation adapterは共通のインターフェースを実装する形にし、
MVPでは組み込みadapterのみを提供するが、将来のplugin化（10章）に
備えてこの境界だけは最初から明確にしておく。

## 6. Test execution model

- 1テスト内の `given → when → then` はステップの因果順序があるため、
  テスト内では逐次実行する（並列化はテスト単位で行う。8章参照）。
- 並列実行を前提にすると、テスト間でのデータ競合・環境状態の競合が
  課題になる。分離レベルの選択肢は次の通り。

| 分離レベル | 概要 | 速度 | 実装コスト | 利用者の負担 |
|---|---|---|---|---|
| テストごとに専用環境（コンテナ）を起動 | 最大の分離だが起動コストが並列度に比例して増える | 遅い | 中 | 低い |
| 共有環境 + データ名前空間規約 | 1回の実行につき環境は1つ、テストごとに一意なID/prefixでデータを分離 | 速い | 低い | 中（一意性を意識した`given`記述が必要） |
| 共有環境 + スナップショット/ロールバック | トランザクション等でテストごとに状態を巻き戻す | 中速 | 高い（adapterごとに対応が必要） | 低い |

**推奨**: MVPは**「共有環境 + データ名前空間規約」**。Furumaiの
コアバリューは高速な繰り返し実行であり、テストごとにコンテナを
起動するアプローチは思想と矛盾する。データ分離は利用者が
`given`ブロックで一意なキーを払い出す規約（例: テストIDをprefixに
利用するヘルパー関数を提供）でカバーする。真に共有状態を書き換える
テストは、8章で述べる `serial` 指定を利用する。

## 7. Environment management model

- テストに必要な周辺基盤（PostgreSQL、Kafka等）はプロジェクト設定に
  宣言し、Furumaiがテスト実行前にcontainer runtime経由で起動する。
- 対象アプリケーション自体はFurumaiの責務外（既に起動している前提）
  であり、Environment Managerが管理するのはあくまで「対象アプリが
  依存する周辺基盤」のみ。
- ヘルスチェック/readiness確認を経てから実際のテスト実行を開始する。
- ライフサイクルスコープは「1回のテスト実行につき起動・終了」を
  デフォルトとしつつ、ローカル開発時の高速化のため、明示的な
  再利用モード（同一コンテナをrun間で使い回す）を用意する
  （Testcontainersのreuse機能に相当）。
- MVPでは、issueで明示されているPostgreSQL・Kafkaを組み込み
  adapterとして提供し、それ以外は「イメージ・ポート・環境変数・
  ヘルスチェックを直接指定できる汎用container spec」で対応する
  escape hatchを用意する。Redis等の追加組み込みadapterは10章の
  将来拡張とする。
- container runtimeはMVPではDocker一本に絞る（Podman等は将来拡張）。

## 8. Parallel / serial execution model

依存関係の表現方法について2案を比較する。

| 案 | 概要 | 表現力 | 理解しやすさ | 実装コスト |
|---|---|---|---|---|
| 粗粒度（suite単位のparallel/serial切り替え） | テストをsuite（グループ）にまとめ、suite単位で並列/直列を指定 | 低い | 高い | 低い |
| 細粒度（`dependsOn`によるDAG） | テスト単位で依存関係を明示し、依存グラフに基づきスケジューリング | 高い | 低い（デバッグ時にグラフを追う必要） | 高い（循環検出、デッドロック対策等） |

**推奨**: MVPは**粗粒度（suite単位の切り替え）**。デフォルトは
suite間・suite内テスト間ともに並列実行、共有可変状態への依存がある
場合のみ、そのsuiteに`serial`指定を付与して直列化する。DAGベースの
細粒度な依存関係表現は、粗粒度モデルの限界が実際の利用で明らかに
なった時点で10章の拡張として検討する。

並列実行数は設定可能な上限値（デフォルトはCPUコア数等の妥当な値）で
制御し、環境（Environment Manager）側のリソース枯渇を防ぐ。

## 9. MVPとして実装すべき範囲

これまでの推奨をまとめると、MVPスコープは以下になる。

**含む**

- Kotlinによる内部DSL（`given`/`when`/`then`ブロック、data classと
  コレクションを用いたparameterized test）
- Stimulus adapter: HTTP request、shell command、DB操作（Postgres）、
  Kafka publish
- Observation adapter: HTTP response、DB state query（Postgres）、
  Kafka message、stdout/exit code、基本的なAssertion API（等価性・
  部分一致・存在確認程度）
- Environment Manager: Docker経由でのPostgres/Kafkaのライフサイクル
  管理、汎用container spec による任意コンテナのescape hatch
- 実行モデル: suite単位のデフォルト並列実行、`serial`指定による
  直列化、共有環境+データ名前空間規約による分離
- CLI（`furumai test`相当）、console reporter、失敗時の非ゼロ終了
  コード

**含まない（将来拡張へ）**

- Plugin architecture（外部adapterの動的ロード）
- Application lifecycle management（対象アプリ自体の起動/停止）
- DAGベースの細粒度依存関係
- Docker以外のcontainer runtime
- リトライ/フレーキーテスト対策
- JUnit XML等のCI向け構造化レポート（console出力のみ）
- 複数DSLホスト言語対応

## 10. 将来的な拡張ポイント

- **Plugin architecture**: Stimulus/Observation adapterの共通
  インターフェース（5章）を外部パッケージとして動的ロード可能にし、
  サードパーティによるadapter追加（gRPC、GraphQL、S3、SQS/RabbitMQ
  等）を可能にする。
- **Application lifecycle management**: 対象アプリ自体の起動/停止を
  Environment Managerの管理対象に含める（docker-compose連携等）。
- **細粒度の依存関係グラフ**: suite単位の粗粒度モデルの限界が
  見えた場合に、`dependsOn`ベースのDAGスケジューリングへ拡張する。
- **複数container runtime対応**: Podman、リモート/Kubernetesベースの
  環境プロビジョニングなど、CI規模拡大に応じた選択肢を追加する。
- **レポート/可観測性の強化**: JUnit XML、Allure等の連携、
  stimulus/observationステップへのOpenTelemetryトレーシング付与
  （分散システムの障害調査を容易にする）。
- **フレーキーテスト対策**: 自動リトライポリシー、flake検知・
  隔離。
- **多言語DSL対応**: Option B（5章）へ段階的に移行し、Kotlin以外の
  言語（TypeScript等）からも同一エンジンを利用できるようにする
  （plugin protocol経由）。これにより「Furumai自身の実装言語と、
  テスト記述言語を分離する」というissueの理念を、テスト対象言語だけ
  でなくテスト記述言語についても拡張する。
- **Contract/Snapshot testing拡張**: `then`ブロックでAPIスキーマ
  適合性検証やレスポンスのスナップショット比較を行うAssertion API
  拡張。

## まとめ

| 項目 | 推奨 | 未決/将来検討 |
|---|---|---|
| テスト記述方式 | 内部DSL（既存言語のtrailing lambda/builder構文） | 具体的なsyntax |
| 技術スタック | Kotlin (JVM) 単一言語モノリス | 配布形態の詳細（GraalVM native-image採用可否） |
| Core architecture | Test Definition / Execution Engine / Stimulus・Observation Adapters / Environment Manager / Reporter の5層 | Plugin protocolの具体設計 |
| Test execution model | 共有環境 + データ名前空間規約 | スナップショット/ロールバック方式への拡張 |
| Environment management | Docker + Postgres/Kafka組み込みadapter + 汎用container spec | Podman等の追加runtime対応 |
| Parallel/serial model | suite単位の粗粒度切り替え | DAGベースの細粒度依存関係 |
| MVPスコープ | 9章参照 | — |

いずれも「合理的な初期案」であり、実装を進める中で得られる知見に
基づき見直すことを前提とする。
