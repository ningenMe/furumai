---
status: open
created: 2026-08-10
related: https://github.com/ningenMe/furumai/issues/1
---

# 0001: Core design direction

GitHub Issue #1「方針」で依頼された、実装着手前の設計整理を行う。

## やること

Issue #1 の要件を前提に、以下を整理したdesign docを作成する。

1. Furumaiのcore concept
2. 想定されるユーザー体験
3. テスト記述方式の候補と比較
4. 技術スタックの候補と比較
5. Core architecture
6. Test execution model
7. Environment management model
8. Parallel / serial execution model
9. MVPとして実装すべき範囲
10. 将来的な拡張ポイント

## 制約

- まだ決まっていない事項(テスト記述言語、DSL構文、内部アーキテクチャ、adapter設計、container runtime、plugin architecture、application lifecycle management、assertion API)を勝手に確定しない
- 複数案を比較した上で、「どの設計がFurumaiの価値を最も強く実現するか」を基準に合理的な案を提示する

## Definition of Done

- 上記10項目を網羅したdesign docが `docs/` 配下に存在する
- 未決事項について複数案の比較と推奨が示されている
