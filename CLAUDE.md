# lazysfn 開発ルール

## 仕様

仕様書は `SPEC.md` を参照すること。

## コーディング規約

- フォーマッタ: `gofmt`
- linter: `golangci-lint`（デフォルトルールセット）
- 命名規則: Go標準（MixedCaps）に従い、Effective Go に準拠すること
- エラーハンドリング: Go のベストプラクティスに準拠（`fmt.Errorf` での `%w` ラップ、エラーの適切な伝播）
- DocString: エクスポートされたシンボルに対して80%以上のカバレッジを満たすこと

## パッケージ構成

```
cmd/lazysfn/main.go
internal/ui/        -- TUI関連
internal/aws/       -- AWS API連携
internal/config/    -- AWS config パース
```

## エージェント権限

- 参照系のコマンド（ファイル読み取り、検索、git status、git log 等）はユーザー確認なしで実行してよい
- 危険なコマンド（`rm`, `git reset --hard`, `git push --force` 等）は必ずユーザー確認を取ること

## Git ルール

- mainブランチへの直接プッシュは禁止。必ずフィーチャーブランチを作成し、PRを経由すること
- コミットメッセージは Conventional Commits に準拠すること（`feat:`, `fix:`, `test:`, `docs:`, `refactor:`, `chore:` 等）

## テスト方針

- テストファースト（TDD）: テストコードを先に書き、その後に実装する
- セッション分離: テストを書くセッションと実装するセッションは別にすること
- テストコードを書くセッションは実装コードを書かない
- 実装するセッションは `SPEC.md` とテストコードを読んで実装する
