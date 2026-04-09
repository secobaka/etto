# etto

> what's next?

Go製のシンプルなCLI todoマネージャー。「えっと、次なにやるんだっけ」から名前をとりました。

## インストール

```
go install github.com/secobaka/etto@latest
```

ソースからビルドする場合:

```
git clone https://github.com/secobaka/etto.git
cd etto
go install .
```

## 使い方

引数なしで未完了タスクを期限順に表示:

```
$ etto
#1   [ ] 牛乳買う  (H)  04/03 15:00
#3   [ ] レポート提出  (M)  04/05 09:00
```

## コマンド

| コマンド | ショートハンド | 説明 |
|---------|-------------|------|
| `add` | `a` | タスク追加 |
| `list` | `l` / `ls` | タスク一覧 |
| `done` | `d` | 完了/未完了の切替 |
| `edit` | `e` | タスク編集 |
| `remove` | `r` / `rm` | タスク削除 |
| `merge` | `m` / `mg` | 重複タスクを統合 |

### add (タスク追加)

```
etto a "牛乳買う"
etto a "レポート提出" -d "2026-04-03 15:00" -p high
```

オプション:
- `--due` (`-d`) - 期限。`2006-01-02 15:04` 形式
- `--priority` (`-p`) - 優先度。`high` / `medium` / `low` (デフォルト: low)

### list (タスク一覧)

```
etto              # 未完了のみ、期限順
etto ls -s priority  # 優先度順
etto ls --all      # 完了済みも含む
```

### done (完了切替)

```
etto d 3          # タスク #3 の完了/未完了を切替
```

### edit (タスク編集)

```
etto e 3 -t "新しいタイトル"
etto e 3 -d "2026-04-05 10:00" -p high
```

指定したフィールドだけ更新されます。

### remove (タスク削除)

```
etto rm 3
```

### merge (タスク統合)

重複したタスクを1つに統合:

```
etto m 1 3              # #1 に統合（#3 は削除）
etto m 1 3 -t "新タイトル"  # タイトルを上書き
```

優先度は高い方、期限は早い方が採用されます。

## Extra コマンド

| コマンド | ショートハンド | 説明 |
|---------|-------------|------|
| `yabai` | `yb` | 期限切れ・期限間近のタスク表示 |
| `momuri` | | 未完了タスクを全削除 |

### yabai (ヤバいタスク表示)

期限切れ・期限間近のタスクを表示:

```
etto yabai           # 24時間以内(デフォルト)
etto yabai -h 48     # 48時間以内
```

```
!!! YABAI !!! 2 task(s) need your attention !!!

OVERDUE:
  #2  レポート提出  (H)  04/01 17:00

Due within 24h:
  #5  牛乳買う  (M)  04/02 18:00  (remaining: 3h22m)
```

### momuri (もう無理)

もう無理なとき:

```
$ etto momuri
3 active task(s) will be gone forever. (done tasks will be kept)
Really momuri? (y/N): y
Removed 3 task(s). You are free now!
```

## データ

タスクは `~/.etto/tasks.json` に保存されます。
