# redfish-trace

`monitor-hw` の Redfish トラバースが BMC に与える負荷を実機で測るためのツール。
setup-hw に埋め込まれた収集ルール (`redfish.Rules`) をそのまま使い、以下のモードで
同じ BMC をトラバースし、HTTP リクエスト 1 本ごとの OpenTelemetry span を記録して
SVG のウォーターフォールに描く。あわせて各モードが生成する Prometheus メトリクスが
一致することを本物の `redfish.Collector` で検証する。

| mode | 内容 |
|---|---|
| `current` | 今の monitor-hw と同じ動き (セッション確認 → `GET /redfish/v1/` → `@odata.id` を辿る逐次 DFS) |
| `exclude` | 提案 1: 追加 Exclude (既定 `/Oem/`) |
| `expand` | 提案 2: コレクションを `?$expand=*($levels=1)` で 1 回の GET にまとめる |
| `improved` | 1 + 2 |

`expand` では、ルールの Metrics Path (`.../PCIeDevices/{device}` など) からコレクションの
パスを推定し、推定漏れは実行時に学習する (2 サイクル目以降に効く)。メンバーへ他の
リンクから先に到達した場合も、親コレクションを先に expand で取る。

このモジュールは setup-hw 本体とは別の Go module (`replace ../../`) で、
本体の go.mod に otel 依存を持ち込まない。

## 使い方

```console
$ cd tools/redfish-trace
$ go build -o redfish-trace .

# 別の実機に対して (iDRAC の RedfishVersion からルールを自動選択)
$ REDFISH_PASSWORD=... ./redfish-trace -host 10.0.0.1 -user root -cycles 2 -interval 10s -svg idrac.svg

# サーバ上で setup-hw の設定をそのまま使う
$ ./redfish-trace -address-file /etc/neco/bmc-address.json -user-file /etc/neco/bmc-user.json

# ルールを明示 / YAML から読む
$ ./redfish-trace -host ... -rule dell_redfish_1.20.1.yml
$ ./redfish-trace -host ... -rule-file ../../redfish/rules/dell_redfish_1.20.1.yml

# BMC なしで動作確認 (collector show 形式のダンプを使う)
$ ./redfish-trace -input-file ../../testdata/redfish-1.20-from-idrac9-v11.json -fake-latency 20ms -interval 0
```

主なオプション:

- `-modes current,improved` 実行するモードと順番 (`current,exclude,expand,improved`)
- `-cycles 2` モードごとのサイクル数。2 回目でセッション再利用・コネクション再利用・学習の効果が見える
- `-interval 5s` サイクル間の待ち (monitor-hw は 60s)
- `-timeout 5s` リクエストごとのタイムアウト (monitor-hw と同じ)
- `-extra-exclude REGEX` `exclude` / `improved` で追加する Exclude (複数可、既定 `/Oem/`)
- `-follow-nextlink` iDRAC は 50 件でページングし `Members@odata.nextLink` を返す。monitor-hw は追わないので既定は追わない
- `-svg FILE` 出力 SVG。`-no-waterfall` で行ごとのウォーターフォールを省く
- `-spans-json FILE` span を JSON (stdouttrace 形式) でも保存
- `-dump-dir DIR` モードごとの収集結果を `-input-file` に使える形式で保存
- `-no-logout` 終了時に Redfish セッションを DELETE しない (既定では削除してセッション枯渇を防ぐ)
- `-v` リクエストごとにログ

## 出力

標準出力にサイクルごとの集計とメトリクス比較:

```
mode       cycle   duration  requests     ok    err expand  inlined      bytes    tls   reused
current    1          2.16s       639    637      2      0        0     1.3MiB      0        0
improved   1          562ms       152    151      1     99      261   574.4KiB      0        0
improved   2          540ms       148    148      0     98      261   572.5KiB      0        0

metrics: current=1038 improved=1038 (identical to current)
```

続けてリクエスト単位のレイテンシ分布 (kind = all / plain / `$expand`) と、サイクルごとに遅かった
リクエスト上位 5 件が出る。`$expand` の p50 が plain の何倍かで、iDRAC 側の展開コストが読める。

```
mode       cycle  kind          n       p50       p90       p99       max
current    2      plain       637     120ms     180ms     450ms     2.1s
improved   2      plain        51     110ms     160ms     300ms     600ms
improved   2      $expand      99     650ms     1.4s      3.2s      4.8s

slowest requests per cycle:
  current    2        2.1s  200 /redfish/v1/...
```

SVG はサイクルごとに 1 パネル。上段はリクエストを 1 本のレーンに並べた帯 (時間軸は全パネル共通)、
下段は 1 リクエスト 1 行のウォーターフォール。色: 緑 = GET 200、紫 = `$expand`、青 = 304、
赤 = non-2xx、暗赤 = エラー/タイムアウト、橙 = セッション操作、茶 = version 取得。
行頭の黒線は TLS ハンドシェイクが発生したリクエスト。バーにマウスを載せるとパス、
ステータス、バイト数、所要時間、コネクション再利用の有無が出る。

span の属性: `url.path`, `url.query`, `http.response.status_code`, `http.response.body.size`,
`redfish.expand`, `redfish.refetch`, `redfish.expanded_members`, `net.conn.reused`,
`tls.handshake`, `tls.handshake.ms`, `redfish.kind`, `redfish.mode`, `redfish.cycle`。
