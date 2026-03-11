# lazysfn

AWS Step Functionsの情報をターミナルから確認できるTUIツール。lazygit風のUI/UXでStep Functionsの稼働状況を閲覧する。

https://github.com/user-attachments/assets/a4ff0e7a-87cb-4f60-b3b0-07efa41ec525

## 機能

- AWS profileの選択（`~/.aws/config` から読み込み）
- ステートマシン一覧の表示（Standardタイプのみ）
- 実行履歴の表示（実行結果、失敗ステート、Input Param等）
- ステータスの色分け表示
- ステートマシン名のインクリメンタル検索
- Vim準拠のキーバインド

## インストール
```
git clone https://github.com/myuron/lazysfn
```

## キーバインド

### グローバル

| キー | 動作 |
|---|---|
| `?` | キーバインドヘルプの表示 / 非表示 |
| `q` | 終了 / ポップアップを閉じる |
| `R` | 更新 |

### メイン画面

| キー | 動作 |
|---|---|
| `j` / `k` | カーソル下 / 上 |
| `h` / `l` | 左パネル / 右パネルにフォーカス |
| `Tab` | パネル切り替え |
| `/` | インクリメンタル検索（左パネル） |

### 検索モード

| キー | 動作 |
|---|---|
| `Esc` | 検索をキャンセル（全件表示に戻る） |
| `Enter` | 検索を確定（フィルタを維持） |

### プロファイル選択

| キー | 動作 |
|---|---|
| `j` / `k` | カーソル下 / 上 |
| `Enter` | プロファイルを選択 |
| `q` | 終了 |

## 開発環境

[Nix](https://nixos.org/) を使用。

```sh
nix develop
```
