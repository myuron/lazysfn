# lazysfn

AWS Step Functionsの情報をターミナルから確認できるTUIツール。lazygit風のUI/UXでStep Functionsの稼働状況を閲覧する。

## 機能

- AWS profileの選択（`~/.aws/config` から読み込み）
- ステートマシン一覧の表示（Standardタイプのみ）
- 実行履歴の表示（実行結果、失敗ステート、Input Param等）
- ステータスの色分け表示
- Vim準拠のキーバインド

## 開発環境

[Nix](https://nixos.org/) を使用。

```sh
nix develop
```
