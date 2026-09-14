# iDRAC 負荷試験 手順書 (redfish-trace)

monitor-hw の Redfish トラバースが iDRAC に与える負荷を、現状 (`current`) と
改善案 (`improved` = Oem 除外 + `$expand`) で比較計測する手順。

前提:

- 対象サーバで monitor-hw (setup-hw コンテナ) は動いていない
- 作業マシンから iDRAC の 443/tcp に到達できる
- iDRAC の読み取り専用ユーザー (`support` 相当) の ID とパスワードがある
- 作業マシンに Go 1.26 以上がある (なければ手順 1 を別マシンで行いバイナリを持ち込む)

## 1. ビルド

```console
cd tools/redfish-trace
CGO_ENABLED=0 go build -o redfish-trace .
./redfish-trace -h 2>&1 | head -5
```

別マシンでビルドして持ち込む場合は、上で出来た `redfish-trace` (静的バイナリ) を
`scp` で配置する。

## 2. 接続情報を環境変数に入れる

パスワードはコマンドラインに書かず、`read -s` で環境変数に入れる。
ツールは `REDFISH_PASSWORD` を自動で読む。以降のコマンドは履歴に残っても問題ない。

```console
export IDRAC_HOST=10.0.0.1      # iDRAC の IP
export IDRAC_USER=support       # 読み取り専用ユーザー
read -rs 'REDFISH_PASSWORD?iDRAC password: ' && export REDFISH_PASSWORD
```

bash の場合は `read -rs -p 'iDRAC password: ' REDFISH_PASSWORD && export REDFISH_PASSWORD`。

作業終了時は `unset REDFISH_PASSWORD` する (手順 8)。

## 3. 疎通確認

```console
mkdir -p result && cd result
../redfish-trace -host "$IDRAC_HOST" -user "$IDRAC_USER" \
  -modes current -cycles 1 -svg smoke.svg -v 2>&1 | tee smoke.log
```

確認すること:

- 先頭の `rule: dell_redfish_x.y.z.yml`。
  `warning: no exact rule for RedfishVersion ...` が出た場合は、そのファームウェア版では
  monitor-hw 本体が収集に失敗する状態なので、その旨を記録する。
- `cycle 1: N resources in T` が出ること。
- `!` で始まる行 (非 200 / タイムアウト) の有無とパス。
- 最後に `wrote smoke.svg`。

ここで `login: status 401` なら ID/パスワードが違う。`context deadline exceeded` が
連発するなら iDRAC が既に重いか、経路に問題がある。

## 4. 本計測

monitor-hw と同じ 60 秒間隔で、各モード 3 サイクル回す。所要時間はおよそ
(トラバース時間 + 60 秒) × 3 × 2 モードで、10〜15 分を見込む。

```console
../redfish-trace -host "$IDRAC_HOST" -user "$IDRAC_USER" \
  -modes current,improved -cycles 3 -interval 60s \
  -svg "idrac-${IDRAC_HOST}.svg" -spans-json "idrac-${IDRAC_HOST}.json" \
  -dump-dir dump 2>&1 | tee "idrac-${IDRAC_HOST}.log"
```

途中で止める場合は Ctrl-C。セッションの DELETE を試みてから終了する。

4 モードすべてを見たい場合は `-modes current,exclude,expand,improved`
(所要時間は約 2 倍)。

## 5. 結果の読み方

ログ末尾の集計表 (行 = モード × サイクル):

```
mode       cycle   duration  requests     ok    err expand  inlined      bytes    tls   reused
current    1          ...        6xx    ...    ...      0        0       ...      1      ...
current    2          ...
improved   1          ...        1xx    ...    ...     9x      2xx       ...      1      ...
improved   2          ...
```

| 列 | 見る点 |
|---|---|
| `duration` | 1 サイクルのトラバース時間。current と improved の差が削減効果。2 サイクル目以降 (定常) で比較する |
| `requests` | HTTP リクエスト数。オフライン再現では current 639、improved 148 |
| `err` | タイムアウト (5 秒) と非 200 の数。current で多いなら iDRAC が既に詰まっている |
| `expand` | `$expand` 付き GET の数。improved で 0 なら iDRAC が `$expand` を受け付けていない |
| `inlined` | `$expand` で 1 回の応答に畳み込まれたメンバー数 |
| `tls` | TLS ハンドシェイク回数。2 サイクル目以降も毎回 1 以上なら、60 秒アイドルで iDRAC が接続を切っている |
| `reused` | コネクションを再利用したリクエスト数。`requests` に近いのが正常 |

その下の `metrics: current=N improved=N (identical to current)` が一致していれば、
改善案でも monitor-hw が出すメトリクスは変わらない。センサ値はサイクル間で変わるので、
`missing`/`extra` に同じメトリクス名で値だけ違うものが出るのは正常。名前ごと消えている
ものがあれば記録する。

SVG はブラウザで開く。サイクルごとに 1 パネルで、上段の帯がサイクル全体、下段が
1 行 1 リクエスト。緑 = 200、紫 = `$expand`、赤 = 非 200、暗赤 = タイムアウト、
行頭の黒線 = TLS ハンドシェイク。バーにマウスを載せると詳細が出る。

判断の目安:

- `$expand` 1 本の所要時間が、置き換えた個別 GET の合計と同程度なら、iDRAC 側の
  コストは内部取得が支配的で、リクエスト数削減の効果は主にラウンドトリップ分。
- improved の `duration` が current の半分以下なら、リクエスト数がボトルネック。

## 6. セッションが残っていないことを確認

ツールは終了時にセッションを DELETE する。念のため残数を確認する。

```console
curl -sk -u "$IDRAC_USER:$REDFISH_PASSWORD" \
  "https://$IDRAC_HOST/redfish/v1/SessionService/Sessions" | jq '."Members@odata.count"'
```

計測前より増えていれば、`Members` の `@odata.id` に対して DELETE する
(ツールの `-no-logout` を使っていない限り通常は残らない)。

## 7. 記録するもの

- `idrac-<IP>.log` (集計表とメトリクス比較)
- `idrac-<IP>.svg`、`idrac-<IP>.json`
- `dump/current.json`、`dump/improved.json` (後で `-input-file dump/current.json` を指定すると
  BMC なしで再現できる)
- サーバの機種、iDRAC ファームウェア版、`rule:` 行に出たルール名
- 計測中に他から iDRAC へアクセスがあったか (Web UI ログイン等)

## 8. 後片付け

```console
unset REDFISH_PASSWORD
cd .. && rm -rf result/dump   # ダンプは BMC の全リソースを含むので共有先に注意
```

## トラブルシュート

| 症状 | 対処 |
|---|---|
| `login: status 401` | ID/パスワードを確認。`read -s` を再実行 |
| `login: status 403` や 500 | iDRAC のセッション上限に達している可能性。手順 6 で残セッションを確認 |
| `warning: no exact rule` | monitor-hw もそのファームウェアでは失敗する。`-rule dell_redfish_1.20.1.yml` を明示して計測は続行できる |
| improved で `expand` が 0 | ルールから推定したコレクションパスが実機に無い。`-v` で `?$expand` 付き GET の status を確認 |
| improved で `err` が増える | `$expand` を拒否するコレクションがある。該当パスを記録 (rule への除外リスト追加の判断材料) |
| 全体に `context deadline exceeded` | `-timeout 10s` にして再実行。それでも多いなら iDRAC 側の問題 |
