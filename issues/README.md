# issues/ ディレクトリについて

このリポジトリはv0が完成するまでprivateで開発するため、タスク管理はGitHub Issuesではなくこのディレクトリで行う。v0完成後、必要に応じてGitHub Issuesに移行する。

## ルール

- 1タスク1ファイル: `issues/NNNN-slug.md`(`NNNN` は4桁の連番、`slug` は英語のkebab-case)
- 各ファイルの先頭にYAML frontmatterを置く

```yaml
---
status: open       # open | in-progress | done
created: 2026-08-10 # YYYY-MM-DD
related:            # 関連するGitHub Issue番号やPR番号(あれば)
---
```

- タスクが完了したら `status: done` に更新する(ファイルは削除せずログとして残す)
- 着手したら `status: in-progress` に更新する
- 次の連番は既存ファイルの最大値+1
