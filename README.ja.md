<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus のロゴ：藍色の正方形に淡い色の二つの葉" width="96"></p>

# Thalamus

[English](https://github.com/lfoletti/thalamus/blob/main/README.md) | [简体中文](https://github.com/lfoletti/thalamus/blob/main/README.zh-CN.md) | 日本語 | [한국어](https://github.com/lfoletti/thalamus/blob/main/README.ko.md) | [Deutsch](https://github.com/lfoletti/thalamus/blob/main/README.de.md)

[![CI](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/lfoletti/thalamus)](https://github.com/lfoletti/thalamus/releases) [![Licence: MIT](https://img.shields.io/github/license/lfoletti/thalamus)](LICENSE)

Thalamus は、組織の知識を各ユーザーとそのコーディングエージェントに、各人の権限の範囲内で届けます。[CortexDB](https://github.com/liliang-cn/cortexdb) の上に構築されています。

> **開発中です。** インターフェースはバージョン間で変わります。

## できること

チームは知識を複数の CortexDB ブレインに保管します。チーム用に一つ、顧客用に一つ、組織全体用に一つ、といった形です。各ユーザーはそのうちのいくつかを読むことができ、書き込めるのはさらに少数です。Thalamus は、ユーザーとこれらのブレインの間にあるゲートウェイです。

ユーザーのアシスタントは、MCP サーバーであるゲートウェイを通じてブレインにアクセスします。アシスタントが `th_recall` で質問すると、ゲートウェイはこのユーザーが読める各ブレインを、ユーザー自身のキーで検索し、結果を統合します。各抜粋には出典が付きます。ブレイン、文書、作成者です。`th_get` を使うと、アシスタントは抜粋の元の文書全体を読みます。

ユーザーは Claude Code から、または Chrome のページから、文書を提案することもできます。文書はブレインのキュレーターを待ち、キュレーターがそれを承認または却下します。

## 使う理由

### 1. ガバナンス

Thalamus は薄いガバナンス層です。組織が知識を落ち着いて効率よく共有できるようにし、各アクセスを管理下に置きます。管理は簡単です。権限は一つのファイルにまとまり、キーは権限に従い、変更は再起動なしで反映されます。

- **申請なしのアクセス。** ユーザーは組織のアカウント（OpenID Connect）でログインします。ゲートウェイは権限が許すキーを発行します。ある人にチームのブレインへのアクセスを与えるには、`rights.yaml` のチームのグループにそのログイン名を追加します。変更は次の呼び出しから、再起動なしで反映されます。
- **1 人 1 ブレインにつき 1 キー。** CortexDB は呼び出しのたびにキーを検証するため、キーの範囲はゲートウェイに依存しません。ある 1 人のアクセスを取り消すとき、ゲートウェイはその人のキーだけを削除します。
- **一つの手順で知識を追加。** ユーザーは Claude Code から、または Chrome のページから文書を提案します。ブレインのキュレーターは、同じツールからそれを承認または却下します。
- **簡単な管理。** `th-gateway key list` はキーと権限を比較し、`key sync` はその差分を修正します。`th-gateway audit` は読み取り、書き込み、拒否、ログインの概要を示します。

### 2. コンプライアンスと主権

Thalamus は、GDPR（EU 一般データ保護規則）、スイス連邦データ保護法（FADP）、そして高リスクの用途については EU AI 法へのコンプライアンスを支援するように作られています。ツールが組織をコンプライアンス適合にするわけではありません。ただし Thalamus は、これらの枠組みが求める管理策を提供します。

| 管理策 | Thalamus での実装 | 条文 |
| :- | :- | :- |
| データの最小化 | 読み取りはブレインごと、日ごとに、名前なしで集計されます。名前付きの読み取り記録を持つのは `audit_reads: true` のブレインだけです。 | GDPR 5条1項(c)、25条2項。FADP 6条2項、7条3項。 |
| 保存の制限 | `audit: retention_days` で監査記録の保存期間を制限します。 | GDPR 5条1項(e)。FADP 6条4項。 |
| セキュリティと追跡可能性 | 個人ごとのキーをエンジンが検証します。書き込み、レビュー、ログイン、拒否された呼び出しは、本人の名前付きで 1 行ずつ記録されます。 | GDPR 5条1項(f)、5条2項、32条。FADP 8条。 |
| データの所在 | 組織がゲートウェイと CortexDB サーバーをホストします。ローカルブレインはユーザーのマシン上にとどまり、ゲートウェイを経由しません。 | GDPR 44条。FADP 16条。 |
| 透明性 | 各抜粋は、そのブレイン、文書、作成者を示します。 | GDPR 5条1項(a)。 |
| **対象外：** モデル提供者への移転 | フックはプロンプトに文章を挿入し、プロンプトはアシスタントのモデル提供者に送られます。Thalamus はこの移転を管理しません。組織が自らこれに対処します。提供者との処理契約、提供者がスイスまたは EU の外にある場合の移転保護措置、そして提供者側での保存設定です。 | GDPR 28条、44条〜46条。FADP 9条、16条。 |

`audit_reads: true` のブレインは名前付きの読み取りログを保持し、そのログは個人データです。組織はユーザーに通知し、アクセス請求に応じ、保存期間を定めます（GDPR 13条、15条。FADP 19条、25条）。

**高リスクの用途の場合。** 以下の AI 法の条文は、高リスク AI システムにのみ適用されます。コーディングエージェント向けのリコールは、一般には高リスクの用途ではありません。高リスクシステムの導入者（デプロイヤー）に対して、Thalamus は対応する管理策を提供します。

| 管理策 | Thalamus での実装 | 条文 |
| :- | :- | :- |
| 記録の保持 | 書き込み、レビュー、ログイン、拒否された呼び出しは、本人の名前付きで 1 行ずつ記録されます。 | AI 法 12条。 |
| 人間による監視 | 提案された文書は、リコールで提供される前にキュレーターを待ちます。 | AI 法 14条、26条2項。 |

GDPR：規則 (EU) 2016/679。FADP：改正スイス連邦データ保護法（SR 235.1）。AI 法：規則 (EU) 2024/1689。

### 3. ローカルブレインと共有ブレインを一つのリコールで

- **適切な大きさのブレイン。** ブレインは CortexDB サーバー全体、または共有サーバーの一つのコレクションです。小さなチームは一つのサーバーを共有できます。機密性の高いブレインは専用のサーバーを持てます。
- **フェデレーション。** 一つの質問が、ユーザーが読めるすべてのブレインに届きます。ゲートウェイはブレインごとの重みを付けて結果を統合します。ユーザーは検索先を選びません。
- **ローカルと共有を同時に。** Claude Code では、フックがプロンプトのたびに、共有ブレインとマシンのローカルブレインに同時に問い合わせます。各人は `/th-recall` で自動リコールのブレインを選びます（[ローカルブレインとグローバルブレインの組み合わせ](#combining-local-and-global-brains)を参照）。

### 4. 確認できる正確さ

- **出典付きの文章。** リコールは短い抜粋ではなく、文書の文章（パッセージ）を提供します。`th_get` を使うと、エージェントは文書全体を読みます。
- **要約のピラミッド。** ブレインは文書の上に要約のコレクションを持てます。`th_recall` はまず要約を提供し、`th_descend` は要約が引用する文書を開きます。
- **測定済み。** ある研究のパイロットでは、自動リコールで見つかった事実の割合が 51.7 % から、パッセージと要約を使うと 81.3 % に、さらに英語での質問の言い換えを加えると 85.1 % になりました。
- **ブレインをまたぐエンティティ。** `th_entity` は、ユーザーが読めるすべてのブレインで一つのエンティティを見つけます。隣接ノードをノード ID で統合し、それぞれを提供するブレインを示します。
- **次の予定。** CortexDB v2.139.0 のツールを使った、要約の各文とその出典との照合。

### 5. 多言語

ユーザーは自分の言語で質問しますが、答えとなる文書は別の言語であることがよくあります。`th_recall` は質問を他の言語でも受け取り（`alternate_queries`）、それぞれを別々に検索します。

- スイスのベンチマーク（`tests/eval_ch/`）には、連邦憲法に関するドイツ語、フランス語、イタリア語、ロマンシュ語の 40 問があります。各質問は、別の言語の条文を求めます。
- `multilingual-e5-large` と言い換えを使うと、期待される条文の 97 % が上位 10 件に入り、ロマンシュ語では 91 % です。エンコーダーなしでは 7 % にとどまります。

コーディングエージェントに対して、Thalamus は知識をセッションに届けます。各アシスタントには MCP を通じて、Claude Code ではプロンプトごとの自動リコールを通じてです。

## 概念

| 用語 | 意味 |
| :- | :- |
| ブレイン | CortexDB サーバー、または共有サーバーの一つのコレクション。 |
| 権限 | ファイル `rights.yaml`。ブレイン、グループ、そして各ブレインを誰が読み、書き、キュレーションするかを宣言します。 |
| キー | 1 人の、1 つのブレインに対する CortexDB の API キー。ゲートウェイはこれをサーバーのキーファイルに書き込み、ブレインのコレクションに限定し、読み取り専用または読み書き可能とします。 |
| キュレーター | ブレインに提案された文書を承認または却下する人。 |

## 仕組み

<p align="center"><img src="docs/img/architecture.svg" alt="Claude Code プラグイン、Chrome 拡張機能、th-gateway login が Thalamus ゲートウェイを呼び出します。ゲートウェイは rights.yaml を読み、audit.jsonl に書き込み、呼び出し元のキーで CortexDB サーバーを呼び出します。キーは 1 人 1 ブレインにつき 1 つです。" width="760"></p>

- ゲートウェイは `rights.yaml` を読み、キーを発行します。ファイルが変更されると読み直すため、削除された権限は次の呼び出しから反映されます。
- 各呼び出しは、呼び出し元のキーで CortexDB に送られます。CortexDB は呼び出しのたびにキーを検証するため、キーの範囲はゲートウェイに依存しません。
- ゲートウェイは、セキュリティと追跡可能性のために `audit.jsonl` に監査記録を残します。書き込み、レビュー、ログイン、拒否された呼び出しは、本人の名前付きで 1 行ずつ記録されます。読み取りはブレインごと、日ごとに、名前なしで集計されます。名前付きの読み取り記録を持つのは、`rights.yaml` で `audit_reads: true` としたブレインだけです。`audit: retention_days` で記録の保存期間を制限します。
- ユーザーは組織の ID プロバイダー（OpenID Connect）でログインします。その後、ゲートウェイは権限が許すキーを発行します。

## <a id="combining-local-and-global-brains"></a>ローカルブレインとグローバルブレインの組み合わせ

Claude Code 用の Thalamus プラグインは、ゲートウェイのツールにフックを加えます。プロンプトのたびに、フックは共有ブレインとマシンのローカルブレインに同時に問い合わせます。エージェントが回答する前に、最良の抜粋をエージェントに渡します。ユーザーがそれを求める必要はありません。

<p align="center"><img src="docs/img/recall-hook.svg" alt="ユーザーがプロンプトを入力します。Thalamus のフックはそれをユーザーのキーとともに Thalamus ゲートウェイの th_recall に送り、ゲートウェイはサーバーのブレインを検索します。同時に、フックはマシン上のローカルブレインを検索します。フックは最良の抜粋を出典付きでプロンプトに加えます。その後、エージェントが回答します。" width="760"></p>

- フック（`th-hook`、`UserPromptSubmit` イベント）は、ユーザーのキーでプロンプトを `th_recall` に送ります。文書の最良のパッセージを出典付きで、9,000 文字以内でプロンプトに加えます。10 件のパッセージで、ブレインに要約のピラミッドがある場合はその前に要約のパッセージ 4 件が入ります。
- 3 語未満のプロンプトや `/` コマンドについては問い合わせません。
- プロンプトを止めることはありません。ゲートウェイが応答しない場合やキーが拒否された場合、プロンプトは抜粋なしで続行します。
- エージェントは `th_get` で抜粋の文書全体を開けます。

- ローカルブレインについては、フックは CortexDB のネイティブのリコールをブレインごとに 1 回実行します。既定では、ローカルブレインはプロジェクトのブレインだけです。設定により、ホームディレクトリのブレインや別のプロジェクトのブレインなど、他のローカルブレインを追加できます。
- ローカルブレインはマシン上にとどまります。ゲートウェイを経由せず、監査記録にも含まれません。
- ブロックには、まず共有ブレインの抜粋が入り、次にローカルブレインごとに一つのセクションが続きます。ローカルのセクションは 3,000 文字以内です。
- 各人は `/th-recall`（`on`、`off`、`only`）で、自動リコールのブレインをローカルと共有の両方について選びます。この選択は権限を変えません。

他のアシスタントは MCP を通じて同じツールを使います。自動リコールにはアシスタント側のフックが必要です。Thalamus は Claude Code 用のフックを用意しています。

## インストール

各リリースには次のファイルがあります。

- `thalamus_<version>_<os>_<arch>.tar.gz`：バイナリ `th-gateway` と `th-hook`。macOS と Linux、amd64 と arm64 用。
- `thalamus-plugin_<version>_<os>_<arch>.tar.gz`：Claude Code プラグイン。`bin/` に `th-hook` を含みます。`claude --plugin-dir thalamus-plugin` で Claude Code を起動します。
- `thalamus-extension_<version>.zip`：Chrome 拡張機能。展開し、`chrome://extensions` を開き、デベロッパーモードをオンにしてから、**パッケージ化されていない拡張機能を読み込む**（Load unpacked）をクリックします。
- `checksums.txt`：各ファイルの SHA-256。

`th-gateway version` と `th-hook version` はバイナリのリリースを表示し、ゲートウェイの `GET /v1/version` は鍵なしでそれを返します。デプロイは期待するリリースと照合して確認します。

Go がある場合は、ソースからバイナリをインストールします。

```sh
go install github.com/lfoletti/thalamus/gateway/cmd/th-gateway@latest
go install github.com/lfoletti/thalamus/gateway/cmd/th-hook@latest
```

ゲートウェイはイメージとしても提供されています。`rights.yaml` とキーファイルを一つのディレクトリに置き、それを `/etc/thalamus` にマウントします。

```sh
docker run -p 8787:8787 -v /srv/thalamus:/etc/thalamus ghcr.io/lfoletti/thalamus:latest
```

監査ファイルは `/var/lib/thalamus` にあります。保持するには、そこにボリュームをマウントします。ゲートウェイには、ブレインごとに CortexDB の gRPC サーバーが必要です（[試してみる](#try-it)を参照）。

## <a id="try-it"></a>試してみる

この手順では、二つのブレインを持つ CortexDB サーバーを一つ起動し、次にゲートウェイ、次にプラグインを起動します。すべての手順をこのリポジトリのルートで行います。

1. CortexDB を `gateway/` の隣の `cortexdb/` にクローンし、バイナリをビルドします。

   ```sh
   git clone https://github.com/liliang-cn/cortexdb.git && (cd cortexdb && git checkout v2.139.0)
   mkdir -p bin
   (cd cortexdb && go build -o ../bin/cortexdb-grpc ./cmd/cortexdb-grpc)
   (cd gateway && go build -o ../bin/th-gateway ./cmd/th-gateway && go build -o ../plugin/bin/th-hook ./cmd/th-hook)
   ```

2. 権限と空のキーファイルを書きます。

   ```yaml
   # rights.yaml
   brains:
     team/acme:
       addr: 127.0.0.1:47998
       keys_file: keys.json
       collection: team-acme
       readers: [group:acme]
       writers: [user:alice]
       curators: [user:carol]
     org/handbook:
       addr: 127.0.0.1:47998
       keys_file: keys.json
       collection: handbook
       readers: [group:acme]
       writers: [user:carol]
   groups:
     acme: [alice, bob, carol]
   ```

   ```sh
   echo '{"keys":[]}' > keys.json
   ```

   キーファイルは `keys_file` が示すパスにあります。相対パスは `th-gateway` を実行するディレクトリが起点です。ゲートウェイはこのファイルにキーを書き込み、サーバーは `-keys` でそれを読みます。

3. キーを発行してから、サーバーを起動します。サーバーは起動時にキーファイルを読みます。

   ```sh
   bin/th-gateway key sync rights.yaml
   bin/cortexdb-grpc -db brain.db -addr 127.0.0.1:47998 -keys keys.json &
   ```

4. ゲートウェイを起動します。

   ```sh
   bin/th-gateway serve rights.yaml --listen 127.0.0.1:8787 &
   ```

5. プラグイン付きで Claude Code を起動します。`key sync` が表示したシークレットのいずれかを使います。

   ```sh
   export TH_GATEWAY=http://127.0.0.1:8787 TH_KEY=<secret>
   claude --plugin-dir ./plugin
   ```

`rights.yaml` に `identity` ブロックがある場合は、手順 5 の export の代わりに `th-gateway login http://127.0.0.1:8787` を使います。詳細は `gateway/README.md`、`plugin/README.md`、`extension/README.md` にあります。

## ロードマップ

**完了**（2026年10月25日、リリース 0.4.0）

- 権限、キー、ブレインごとの MCP エンドポイント。
- 人が読めるブレインを対象にしたフェデレーテッドリコール。各抜粋に出典付き。
- 文書の提案と、キュレーターによるレビュー。
- MCP を使わないクライアント向けの REST エンドポイント。
- コマンドラインと拡張機能からの、OpenID Connect によるログイン。
- Claude Code プラグイン：リコールのフック、ツール、`/th-add` コマンド。
- Chrome 拡張機能：ページまたは選択範囲を提案します。ページの主要な内容を残します。
- 比例的な監査：名前付きの書き込み、レビュー、ログイン、拒否。名前なしで集計される読み取り。名前付きの読み取りはそれを求めるブレインのみ。そして保存期間。
- ローカルブレインと共有ブレインを一つのリコールで：フックがマシンのローカルブレインを加え、各人が自動リコールのブレインを選びます。
- 短い抜粋ではなく、文書のパッセージに対するリコール。
- ブレインごとの要約のピラミッド：`th_recall` はまず要約を提供し、`th_descend` は要約が引用する文書を開きます。
- ブレインをまたぐエンティティ：`th_entity` は、各ブレインに対してそのブレイン用のユーザーのキーで一つのエンティティとその隣接ノードを問い合わせ、ノードを ID で統合します。
- `th_recall` または `th_entity` の実行中にブレインが拒否したキーごとの監査記録。
- 複数の言語での質問：`th_recall` は質問を他の言語でも受け取り（`alternate_queries`）、それぞれを別々に検索します。スイスのベンチマークが、ある言語から別の言語（フランス語、ドイツ語、イタリア語、ロマンシュ語）へのリコールを `tests/eval_ch/` で測定します。
- リコールの期限：`th_recall` は `deadline_ms` を受け取り、準備のできたブレインの回答を統合します。フックがこれを使うため、遅いブレインが他のブレインのリコールを止めることはなくなりました。
- Claude Code の中のブレイン：`/brains` はブレインのダイアログを開き、各ブレインが応答するかどうかを示します。Enter で、ブレインを自動リコールに対してオンまたはオフにします。引数なしの `/th-recall` はチェックリストで尋ねます。プロジェクトの `CLAUDE.md` は、`th_add` が書き込むブレインを指定します。
- リリース：macOS と Linux 用の `th-gateway`、`th-hook`、プラグイン、拡張機能のアーカイブ、Docker イメージ `ghcr.io/lfoletti/thalamus`、そして `go install`。[リリース](https://github.com/lfoletti/thalamus/releases)を参照してください。

**CortexDB とともに**

Thalamus は、すべての CortexDB ユーザーに役立つ仕組みについては CortexDB に依拠します。これらの部分は、小さなプルリクエストとして CortexDB に送られます。

- マージ済み：限定されたキー向けのツールの許可リスト（#72〜#74、#78）、コンテンツハッシュと重複チェック（#75）、有効期限（#76）、レビューステータス（#77）、その他の修正（#61、#67〜#71、#82、#83）。
- v2.139.0 でマージ済み、ディスカッション #86 から：
  - キーと署名：サーバーはキーファイルが変更されると読み直すため、新しいキーや取り消されたキーが再起動なしで反映されます（#87）。各書き込みは呼び出し元のキーで署名されます（#88）。メモリは、呼び出し元の範囲が許す場合にのみ読み取られます（#89）。
  - 一つのコレクションに限定されたキー向けのグラフツール（#94〜#98）。`th_entity` はこれを使い、パッセージに加えてブレインをまたいでエンティティを統合します。
  - 三つのツールと一つのインデックス：リクエストログ内の操作対象 ID（#90）、ID のリストに対する `knowledge_get`（#91）、文とその引用元との照合（#92）、そして冗長な二つのエッジインデックスの削除（#93）。
  - ヨーロッパの言語での検索：スイスの評価セット（#99）、カバレッジ下限でのアクセント記号の扱い（#100）、代替クエリの語と実行（#101、#102）。
- 提案中：`upsert_entities` が、自身が指す名前を持つチャンクのコレクションを記録します（#105）。これにより `th_entity` は、Thalamus の外で構築されたブレインのエンティティも提供できます。
- Thalamus は CortexDB v2.139.0 を使用しています。
- 準備中：レビュー待ちの文書を一覧表示するツール。

**次の予定**

- `th_add` で内容を追加したブレインのエンティティ：ゲートウェイにはまだ抽出器がありません。
- `th-hook` の簡単なインストール手段と、Chrome ウェブストアでの拡張機能の公開。
- README での短いデモ。

## 原則

1. **エンジンがすべてのキーを検証します。** ゲートウェイは誰がキーを得るかを決めます。CortexDB は各キーが何をしてよいかを決めます。
2. **1 人 1 ブレインにつき 1 キー。** ゲートウェイは、1 つのブレインに対する 1 人のアクセスを取り消せます。
3. **各抜粋は出典を示します。** 読み手は `th_get` で抜粋の元の文書を開けます。
4. **汎用的な仕組みはアップストリームへ。** Thalamus が持つのは組織のポリシーだけです。権限、グループ、キュレーター、サーバー間のフェデレーションです。

## 開発

変更の前に [CONTRIBUTING.md](CONTRIBUTING.md) を、不具合の報告には [SECURITY.md](SECURITY.md) を、そして [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) をお読みください。

```sh
cd gateway && go test -race ./...           # coverage 100 % outside the two main functions
python3 -I mutate.py                         # one hand mutation per rule, each must fail a test
cd ../extension && npm install && node --test
cd .. && sh tests/e2e_gateway.sh             # a real CortexDB server and the gateway
sh tests/e2e_extension.sh                    # the extension in Chromium
```

各テストはそのコードより先に書かれ、コードなしでは失敗します。手順の全体は `docs/pipeline_pr.md` にあります。プロジェクトの文書の大半はフランス語です。

## ライセンス

MIT。CortexDB も MIT ライセンスです。

---

このファイルは [README.md](https://github.com/lfoletti/thalamus/blob/main/README.md) の翻訳です。内容が異なる場合は英語版が優先されます。issue は英語でお書きください。
