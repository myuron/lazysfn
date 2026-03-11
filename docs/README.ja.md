# lazysfn

AWS Step Functionsの情報をターミナルから確認できるTUIツール。lazygit風のUI/UXでStep Functionsの稼働状況を閲覧する。

[https://github.com/user-attachments/assets/8f76fb05-b507-4118-8d61-6bbaa8145417](https://github.com/user-attachments/assets/39484e9b-fa0a-4abf-abaa-779c115b5f41)

[English README](../README.md)

## 機能

- AWS profileの選択（`~/.aws/config` から読み込み）
- ステートマシン一覧の表示（Standardタイプのみ、名前ソート）
- 直近の実行ステータスを色付き記号 `●` で表示
- 実行履歴の表示（実行ID、実行結果、失敗ステート、開始/終了時間、動作時間、Input Param）
- ステータスの色分け表示（SUCCEEDED: 緑、FAILED: 赤、RUNNING: 青、TIMED_OUT: 黄、ABORTED: グレー）
- ステートマシン名のインクリメンタル検索
- キーバインドヘルプの表示
- 手動リフレッシュ
- エラーモーダル表示（AWS接続エラー時、profile選択への復帰）
- Vim準拠のキーバインド

## 技術スタック

- 言語: Go
- TUIライブラリ: [gocui](https://github.com/jroimartin/gocui)
- AWS SDK: [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)

## インストール

### ビルド

```sh
git clone https://github.com/myuron/lazysfn
cd lazysfn
go build -o ./dist/lazysfn ./cmd/lazysfn
```

### Nix

```sh
nix flake run .#build
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
| 文字入力 | 検索文字列を更新し、リアルタイムで絞り込み |
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
