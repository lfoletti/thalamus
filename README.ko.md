<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus 로고: 남색 정사각형 위의 옅은 엽 두 개" width="96"></p>

# Thalamus

[English](https://github.com/lfoletti/thalamus/blob/main/README.md) | [简体中文](https://github.com/lfoletti/thalamus/blob/main/README.zh-CN.md) | [日本語](https://github.com/lfoletti/thalamus/blob/main/README.ja.md) | 한국어 | [Deutsch](https://github.com/lfoletti/thalamus/blob/main/README.de.md)

[![CI](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/lfoletti/thalamus)](https://github.com/lfoletti/thalamus/releases) [![Licence: MIT](https://img.shields.io/github/license/lfoletti/thalamus)](LICENSE)

Thalamus는 조직의 지식을 각 사용자와 그 사용자의 코딩 에이전트에게, 각자의 권한 범위 안에서 전달합니다. [CortexDB](https://github.com/liliang-cn/cortexdb)를 기반으로 만들어졌습니다.

> **개발 중입니다.** 인터페이스는 버전마다 바뀝니다.

## 하는 일

팀은 지식을 여러 CortexDB 브레인에 보관합니다. 팀용 브레인 하나, 고객용 브레인 하나, 조직 전체용 브레인 하나가 있는 식입니다. 각 사용자는 그중 일부를 읽을 수 있고, 쓸 수 있는 브레인은 그보다 적습니다. Thalamus는 사용자와 이 브레인들 사이의 게이트웨이입니다.

사용자의 어시스턴트는 MCP 서버인 게이트웨이를 통해 브레인에 접근합니다. 어시스턴트가 `th_recall`로 질문하면, 게이트웨이는 이 사용자가 읽을 수 있는 각 브레인을 사용자 본인의 키로 검색하고 결과를 병합합니다. 각 발췌에는 출처가 붙습니다. 출처는 브레인, 문서, 작성자입니다. 어시스턴트는 `th_get`으로 발췌 뒤에 있는 문서 전체를 읽습니다.

사용자는 Claude Code에서, 또는 Chrome의 페이지에서 문서를 제안할 수도 있습니다. 문서는 브레인의 큐레이터를 기다리며, 큐레이터가 이를 승인하거나 거부합니다.

## 사용하는 이유

### 1. 거버넌스

Thalamus는 얇은 거버넌스 계층입니다. 조직이 지식을 차분하고 효율적으로 공유할 수 있게 하며, 각 접근을 통제 아래 둡니다. 관리는 간단합니다. 권한은 파일 하나에 있고, 키는 권한을 따르며, 변경은 재시작 없이 적용됩니다.

- **요청 없는 접근.** 사용자는 조직 계정(OpenID Connect)으로 로그인합니다. 게이트웨이는 권한이 허용하는 키를 발급합니다. 어떤 사람에게 팀의 브레인 접근 권한을 주려면 `rights.yaml`에서 팀 그룹에 그 사람의 로그인을 추가합니다. 변경은 다음 호출부터 재시작 없이 적용됩니다.
- **사람마다, 브레인마다 키 하나.** CortexDB가 모든 호출에서 키를 검사하므로, 키의 범위는 게이트웨이에 의존하지 않습니다. 한 사람의 접근을 회수할 때 게이트웨이는 그 사람의 키만 삭제합니다.
- **한 단계로 추가되는 지식.** 사용자는 Claude Code에서, 또는 Chrome의 페이지에서 문서를 제안합니다. 브레인의 큐레이터는 같은 도구에서 이를 승인하거나 거부합니다.
- **간단한 관리.** `th-gateway key list`는 키를 권한과 비교하고, `key sync`는 그 차이를 바로잡습니다. `th-gateway audit`는 읽기, 쓰기, 거부, 로그인의 요약을 보여 줍니다.

### 2. 규정 준수와 주권

Thalamus는 GDPR, 스위스 연방 데이터 보호법, 그리고 고위험 용도의 경우 EU 인공지능법(AI Act) 준수를 지원하도록 만들어졌습니다. 도구가 조직을 규정 준수 상태로 만들어 주지는 않습니다. 다만 Thalamus는 이 규범들이 요구하는 통제 수단을 제공합니다.

| 통제 | Thalamus에서 | 조항 |
| :- | :- | :- |
| 데이터 최소화 | 읽기는 브레인별, 일별로 이름 없이 집계됩니다. `audit_reads: true`가 설정된 브레인만 실명 읽기 기록을 남깁니다. | GDPR 5(1)(c), 25(2). FADP 6(2), 7(3). |
| 보관 기간 제한 | `audit: retention_days`는 감사 기록이 보관되는 기간을 제한합니다. | GDPR 5(1)(e). FADP 6(4). |
| 보안과 추적성 | 엔진이 검사하는 개인 키. 쓰기, 검토, 로그인, 거부된 호출에는 해당 사람의 이름이 적힌 기록이 남습니다. | GDPR 5(1)(f), 5(2), 32. FADP 8. |
| 데이터 위치 | 조직이 게이트웨이와 CortexDB 서버를 호스팅합니다. 로컬 브레인은 사용자의 컴퓨터에 머물며 게이트웨이를 거치지 않습니다. | GDPR 44. FADP 16. |
| 투명성 | 각 발췌는 브레인, 문서, 작성자를 표시합니다. | GDPR 5(1)(a). |
| **다루지 않는 부분:** 모델 제공자로의 이전 | 훅은 프롬프트에 구절을 넣고, 프롬프트는 어시스턴트의 모델 제공자에게 전송됩니다. Thalamus는 이 이전을 통제하지 않습니다. 조직이 직접 이를 다룹니다. 제공자와의 처리 위탁 계약, 제공자가 스위스나 EU 밖에 있을 때의 이전 보호 조치, 제공자 측의 보관 설정이 이에 해당합니다. | GDPR 28, 44~46. FADP 9, 16. |

`audit_reads: true`가 설정된 브레인은 실명 읽기 로그를 보관하며, 이 로그는 개인정보입니다. 조직은 사용자에게 이를 알리고, 열람 요청에 응하며, 보관 기간을 정합니다(GDPR 13, 15. FADP 19, 25).

**고위험 용도인 경우.** 아래의 AI Act 조항은 고위험 AI 시스템에만 적용됩니다. 코딩 에이전트를 위한 리콜은 일반적으로 고위험 용도가 아닙니다. 고위험 시스템의 배포자에게 Thalamus는 다음과 같은 통제 수단을 제공합니다.

| 통제 | Thalamus에서 | 조항 |
| :- | :- | :- |
| 기록 보관 | 쓰기, 검토, 로그인, 거부된 호출에는 해당 사람의 이름이 적힌 기록이 남습니다. | AI Act 12. |
| 인간의 감독 | 제안된 문서는 리콜이 이를 제공하기 전에 큐레이터를 기다립니다. | AI Act 14, 26(2). |

GDPR: EU 일반 개인정보 보호법(규정 (EU) 2016/679). FADP: 개정 스위스 연방 데이터 보호법(revDSG, SR 235.1). AI Act: EU 인공지능법(규정 (EU) 2024/1689).

### 3. 하나의 리콜 안의 로컬 브레인과 공유 브레인

- **알맞은 크기의 브레인.** 브레인은 CortexDB 서버 전체이거나, 공유 서버의 컬렉션 하나입니다. 작은 팀들은 서버 하나를 함께 쓸 수 있습니다. 민감한 브레인은 전용 서버를 가질 수 있습니다.
- **연합.** 질문 하나가 사용자가 읽을 수 있는 모든 브레인에 전달됩니다. 게이트웨이는 브레인별 가중치를 적용해 결과를 병합합니다. 사용자는 어디를 찾을지 고르지 않습니다.
- **로컬과 공유를 함께.** Claude Code에서는 훅이 프롬프트마다 공유 브레인과 컴퓨터의 로컬 브레인에 동시에 질의합니다. 각 사람은 `/th-recall`로 자동 리콜의 브레인을 고릅니다([로컬 브레인과 글로벌 브레인의 결합](#combining-local-and-global-brains) 참조).

### 4. 확인할 수 있는 정확성

- **출처가 붙은 구절.** 리콜은 짧은 발췌가 아니라 문서의 구절을 제공합니다. 에이전트는 `th_get`으로 문서 전체를 읽습니다.
- **종합의 피라미드.** 브레인은 문서 위에 종합(synthesis) 컬렉션을 둘 수 있습니다. `th_recall`은 종합을 먼저 제공하고, `th_descend`는 종합이 인용하는 문서를 엽니다.
- **측정 결과.** 한 연구 파일럿에서 자동 리콜이 찾은 사실의 비율은 구절과 종합을 사용했을 때 51.7 %에서 81.3 %로, 질문의 영어 변형을 더했을 때 85.1 %로 올랐습니다.
- **브레인을 가로지르는 엔티티.** `th_entity`는 사용자가 읽을 수 있는 모든 브레인에서 하나의 엔티티를 찾습니다. 이웃 노드를 노드 id로 병합하고, 각 이웃을 제공하는 브레인을 표시합니다.
- **다음 단계.** CortexDB v2.139.0의 도구를 사용해 종합의 각 문장을 그 출처와 대조하는 검사.

### 5. 다국어

사용자는 자신의 언어로 질문하고, 답이 되는 문서는 다른 언어로 되어 있는 경우가 많습니다. `th_recall`은 다른 언어로 된 질문(`alternate_queries`)을 받아 각각 따로 검색합니다.

- 스위스 벤치(`tests/eval_ch/`)에는 연방 헌법에 관한 질문 40개가 독일어, 프랑스어, 이탈리아어, 로만슈어로 들어 있습니다. 각 질문은 다른 언어로 된 조문을 요구합니다.
- `multilingual-e5-large`와 변형 질문을 사용하면 기대한 조문의 97 %가 상위 10개 결과 안에 있고, 로만슈어는 91 %입니다. 인코더가 없으면 7 %만 들어 있습니다.

코딩 에이전트에게 Thalamus는 지식을 세션 안으로 가져옵니다. 모든 어시스턴트에는 MCP를 통해, Claude Code에서는 프롬프트마다 실행되는 자동 리콜을 통해 가져옵니다.

## 개념

| 용어 | 의미 |
| :- | :- |
| 브레인 | CortexDB 서버, 또는 공유 서버의 컬렉션 하나. |
| 권한 | `rights.yaml` 파일. 브레인, 그룹, 그리고 각 브레인을 누가 읽고 쓰고 큐레이션하는지를 선언합니다. |
| 키 | 한 브레인에 대한 한 사람의 CortexDB API 키. 게이트웨이는 이를 서버의 키 파일에 기록하며, 키는 브레인의 컬렉션으로 한정되고 읽기 전용 또는 읽기-쓰기입니다. |
| 큐레이터 | 브레인에 제안된 문서를 승인하거나 거부하는 사람. |

## 작동 방식

<p align="center"><img src="docs/img/architecture.svg" alt="Claude Code 플러그인, Chrome 확장 프로그램, th-gateway login이 Thalamus 게이트웨이를 호출합니다. 게이트웨이는 rights.yaml을 읽고 audit.jsonl에 기록하며, 호출자의 키로 CortexDB 서버를 호출합니다. 키는 사람마다, 브레인마다 하나입니다." width="760"></p>

- 게이트웨이는 `rights.yaml`을 읽고 키를 발급합니다. 파일이 바뀌면 다시 읽으므로, 삭제된 권한은 다음 호출부터 적용됩니다.
- 각 호출은 호출자의 키로 CortexDB에 전달됩니다. CortexDB가 모든 호출에서 키를 검사하므로, 키의 범위는 게이트웨이에 의존하지 않습니다.
- 게이트웨이는 보안과 추적성을 위해 `audit.jsonl`에 감사 기록을 남깁니다. 쓰기, 검토, 로그인, 거부된 호출에는 해당 사람의 이름이 적힌 기록이 남습니다. 읽기는 브레인별, 일별로 이름 없이 집계됩니다. `rights.yaml`에서 `audit_reads: true`가 설정된 브레인만 실명 읽기 기록을 남깁니다. `audit: retention_days`는 기록이 보관되는 기간을 제한합니다.
- 사용자는 조직의 ID 제공자(OpenID Connect)로 로그인합니다. 그러면 게이트웨이가 권한이 허용하는 키를 발급합니다.

## <a id="combining-local-and-global-brains"></a>로컬 브레인과 글로벌 브레인의 결합

Claude Code용 Thalamus 플러그인은 게이트웨이의 도구에 훅을 더합니다. 훅은 프롬프트마다 공유 브레인과 컴퓨터의 로컬 브레인에 동시에 질의합니다. 그리고 에이전트가 답하기 전에 가장 좋은 발췌를 에이전트에게 전달합니다. 사용자가 이를 요청할 필요는 없습니다.

<p align="center"><img src="docs/img/recall-hook.svg" alt="사용자가 프롬프트를 입력합니다. Thalamus 훅은 사용자의 키로 이를 Thalamus 게이트웨이의 th_recall에 보내고, 게이트웨이는 서버의 브레인을 검색합니다. 동시에 훅은 컴퓨터의 로컬 브레인을 검색합니다. 훅은 가장 좋은 발췌를 출처와 함께 프롬프트에 추가합니다. 그다음 에이전트가 답합니다." width="760"></p>

- 훅(`th-hook`, `UserPromptSubmit` 이벤트)은 사용자의 키로 프롬프트를 `th_recall`에 보냅니다. 문서의 가장 좋은 구절을 출처와 함께 9,000자 이내로 프롬프트에 추가합니다. 구절은 10개이며, 브레인에 종합의 피라미드가 있으면 종합 구절 4개 뒤에 옵니다.
- 3단어 미만의 프롬프트나 `/` 명령에 대해서는 질의하지 않습니다.
- 프롬프트를 멈추는 일은 없습니다. 게이트웨이가 응답하지 않거나 키가 거부되면, 프롬프트는 발췌 없이 계속 진행됩니다.
- 에이전트는 `th_get`으로 발췌의 문서 전체를 열 수 있습니다.

- 로컬 브레인에 대해서는 훅이 브레인마다 한 번씩 CortexDB의 네이티브 리콜을 실행합니다. 기본적으로 로컬 브레인은 프로젝트의 브레인 하나뿐입니다. 설정으로 다른 로컬 브레인을 추가할 수 있습니다. 예를 들어 홈 디렉터리의 브레인이나 다른 프로젝트의 브레인입니다.
- 로컬 브레인은 컴퓨터에 머뭅니다. 게이트웨이를 거치지 않으며 감사 기록에도 남지 않습니다.
- 블록에는 공유 발췌가 먼저 오고, 그다음 로컬 브레인마다 섹션이 하나씩 옵니다. 로컬 섹션은 3,000자 이내를 사용합니다.
- 각 사람은 `/th-recall`(`on`, `off`, `only`)로 자동 리콜의 브레인을 로컬과 공유 모두에서 고릅니다. 이 선택은 권한을 바꾸지 않습니다.

다른 어시스턴트는 MCP를 통해 같은 도구를 사용합니다. 자동 리콜에는 어시스턴트 쪽의 훅이 필요합니다. Thalamus에는 Claude Code용 훅이 있습니다.

## 설치

각 릴리스에는 다음 파일이 있습니다.

- `thalamus_<version>_<os>_<arch>.tar.gz`: macOS와 Linux, amd64와 arm64용 바이너리 `th-gateway`와 `th-hook`.
- `thalamus-plugin_<version>_<os>_<arch>.tar.gz`: `bin/`에 `th-hook`이 들어 있는 Claude Code 플러그인. `claude --plugin-dir thalamus-plugin`으로 Claude Code를 시작합니다.
- `thalamus-extension_<version>.zip`: Chrome 확장 프로그램. 압축을 풀고 `chrome://extensions`를 연 다음, 개발자 모드를 켜고 **압축해제된 확장 프로그램을 로드합니다**를 클릭합니다.
- `checksums.txt`: 각 파일의 SHA-256.

`th-gateway version`과 `th-hook version`은 바이너리의 릴리스를 출력하고, 게이트웨이의 `GET /v1/version`은 키 없이 그것을 답합니다. 배포는 기대하는 릴리스와 대조하여 확인합니다.

Go가 있으면 소스에서 바이너리를 설치합니다.

```sh
go install github.com/lfoletti/thalamus/gateway/cmd/th-gateway@latest
go install github.com/lfoletti/thalamus/gateway/cmd/th-hook@latest
```

게이트웨이는 이미지로도 제공됩니다. `rights.yaml`과 키 파일을 한 디렉터리에 넣은 다음, 이를 `/etc/thalamus`에 마운트합니다.

```sh
docker run -p 8787:8787 -v /srv/thalamus:/etc/thalamus ghcr.io/lfoletti/thalamus:latest
```

감사 파일은 `/var/lib/thalamus`에 있습니다. 이를 보존하려면 그 위치에 볼륨을 마운트합니다. 게이트웨이에는 브레인마다 CortexDB gRPC 서버가 필요합니다([사용해 보기](#try-it) 참조).

## <a id="try-it"></a>사용해 보기

이 단계들은 브레인 두 개를 가진 CortexDB 서버 하나를 시작하고, 이어서 게이트웨이, 그다음 플러그인을 시작합니다. 모든 단계는 이 저장소의 루트에서 실행합니다.

1. CortexDB를 `gateway/` 옆의 `cortexdb/`에 클론한 다음 바이너리를 빌드합니다.

   ```sh
   git clone https://github.com/liliang-cn/cortexdb.git && (cd cortexdb && git checkout v2.139.0)
   mkdir -p bin
   (cd cortexdb && go build -o ../bin/cortexdb-grpc ./cmd/cortexdb-grpc)
   (cd gateway && go build -o ../bin/th-gateway ./cmd/th-gateway && go build -o ../plugin/bin/th-hook ./cmd/th-hook)
   ```

2. 권한과 빈 키 파일을 작성합니다.

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

   키 파일은 `keys_file`이 지정하는 경로에 있습니다. 상대 경로는 `th-gateway`가 실행되는 디렉터리를 기준으로 합니다. 게이트웨이는 이 파일에 키를 기록하고, 서버는 `-keys`로 이 파일을 읽습니다.

3. 키를 발급한 다음 서버를 시작합니다. 서버는 시작할 때 키 파일을 읽습니다.

   ```sh
   bin/th-gateway key sync rights.yaml
   bin/cortexdb-grpc -db brain.db -addr 127.0.0.1:47998 -keys keys.json &
   ```

4. 게이트웨이를 시작합니다.

   ```sh
   bin/th-gateway serve rights.yaml --listen 127.0.0.1:8787 &
   ```

5. 플러그인과 함께 Claude Code를 시작합니다. `key sync`가 출력한 시크릿 중 하나를 사용합니다.

   ```sh
   export TH_GATEWAY=http://127.0.0.1:8787 TH_KEY=<secret>
   claude --plugin-dir ./plugin
   ```

`rights.yaml`에 `identity` 블록이 있으면, `th-gateway login http://127.0.0.1:8787`이 5단계의 export를 대신합니다. 자세한 내용은 `gateway/README.md`, `plugin/README.md`, `extension/README.md`에 있습니다.

## 로드맵

**완료** (2026년 10월 25일, 릴리스 0.4.0)

- 권한, 키, 브레인마다 MCP 엔드포인트 하나.
- 한 사람이 읽을 수 있는 브레인들에 대한 연합 리콜. 각 발췌에 출처가 붙습니다.
- 제안된 문서와 큐레이터의 검토.
- MCP가 없는 클라이언트를 위한 REST 엔드포인트.
- 명령줄과 확장 프로그램에서의 OpenID Connect 로그인.
- Claude Code 플러그인: 리콜 훅, 도구, `/th-add` 명령.
- Chrome 확장 프로그램: 페이지나 선택 영역을 제안합니다. 페이지의 주요 콘텐츠를 유지합니다.
- 비례적인 감사: 실명 쓰기, 검토, 로그인, 거부 기록, 이름 없이 집계되는 읽기, 요청한 브레인에서만 남는 실명 읽기, 그리고 보관 기간.
- 하나의 리콜 안의 로컬 브레인과 공유 브레인: 훅이 컴퓨터의 로컬 브레인을 추가하고, 각 사람이 자동 리콜의 브레인을 고릅니다.
- 짧은 발췌가 아니라 문서의 구절에 대한 리콜.
- 브레인마다 종합의 피라미드: `th_recall`은 종합을 먼저 제공하고, `th_descend`는 종합이 인용하는 문서를 엽니다.
- 브레인을 가로지르는 엔티티: `th_entity`는 각 브레인에 해당 브레인용 사용자 키로 엔티티 하나와 그 이웃을 질의하고, 노드를 id로 병합합니다.
- `th_recall` 또는 `th_entity` 중에 브레인이 거부한 각 키에 대한 감사 기록.
- 여러 언어로 된 질문: `th_recall`은 다른 언어로 된 질문(`alternate_queries`)을 받아 각각 따로 검색합니다. 스위스 벤치가 `tests/eval_ch/`에서 한 언어에서 다른 언어로의 리콜(프랑스어, 독일어, 이탈리아어, 로만슈어)을 측정합니다.
- 리콜의 마감 시간: `th_recall`은 `deadline_ms`를 받아 준비된 브레인들의 답을 융합합니다. 훅이 이를 사용하므로, 느린 브레인이 더 이상 다른 브레인의 리콜을 멈추지 않습니다.
- Claude Code 안의 브레인: `/brains`는 각 브레인의 응답 여부를 보여 주는 브레인 대화 상자를 열고, Enter로 자동 리콜에서 브레인을 켜거나 끕니다. 인수 없는 `/th-recall`은 체크리스트로 묻습니다. 프로젝트의 `CLAUDE.md`는 `th_add`가 기록할 브레인을 지정합니다.
- 릴리스: macOS와 Linux용 `th-gateway`, `th-hook`, 플러그인, 확장 프로그램의 아카이브, Docker 이미지 `ghcr.io/lfoletti/thalamus`, 그리고 `go install`. [릴리스](https://github.com/lfoletti/thalamus/releases)를 참조하십시오.

**CortexDB와 함께**

Thalamus는 모든 CortexDB 사용자에게 도움이 되는 각 메커니즘을 CortexDB에 맡깁니다. 이러한 부분은 작은 pull request로 CortexDB에 보내집니다.

- 병합됨: 한정된 키를 위한 도구 허용 목록(#72~#74, #78), 콘텐츠 해시와 중복 검사(#75), 만료일(#76), 검토 상태(#77), 그 밖의 수정(#61, #67~#71, #82, #83).
- v2.139.0에서 병합됨, 토론 #86에서 출발:
  - 키와 서명: 서버가 키 파일이 바뀌면 다시 읽으므로, 새 키나 회수된 키가 재시작 없이 적용됩니다(#87). 각 쓰기는 호출자의 키로 서명됩니다(#88). 메모리는 호출자의 범위가 허용할 때만 읽힙니다(#89).
  - 하나의 컬렉션으로 한정된 키를 위한 그래프 도구(#94~#98). `th_entity`는 이를 사용해 구절에 더해 브레인을 가로질러 엔티티를 병합합니다.
  - 도구 세 개와 인덱스 하나: 요청 로그의 접근된 id(#90), id 목록에 대한 `knowledge_get`(#91), 문장을 인용된 출처와 대조하는 검사(#92), 중복된 엣지 인덱스 두 개의 제거(#93).
  - 유럽 언어 검색: 스위스 평가 세트(#99), 커버리지 하한에서의 악센트 처리(#100), 대체 질의의 단어와 실행(#101, #102).
- 제안됨: `upsert_entities`가 자신이 지정하는 청크의 컬렉션을 기록하여(#105), `th_entity`가 Thalamus 밖에서 만들어진 브레인의 엔티티를 제공할 수 있게 합니다.
- Thalamus는 CortexDB v2.139.0을 사용합니다.
- 준비됨: 검토를 기다리는 문서를 나열하는 도구.

**다음**

- `th_add`로 채워진 브레인의 엔티티: 게이트웨이에는 아직 추출기가 없습니다.
- `th-hook`의 간단한 설치 경로, 그리고 Chrome 웹 스토어의 확장 프로그램.
- README의 짧은 시연.

## 원칙

1. **엔진이 모든 키를 검사합니다.** 게이트웨이는 누가 키를 받을지 결정합니다. CortexDB는 각 키가 무엇을 할 수 있는지 결정합니다.
2. **사람마다, 브레인마다 키 하나.** 게이트웨이는 한 브레인에 대한 한 사람의 접근을 회수할 수 있습니다.
3. **각 발췌는 출처를 보여 줍니다.** 읽는 사람은 `th_get`으로 발췌 뒤에 있는 문서를 열 수 있습니다.
4. **일반적인 메커니즘은 업스트림으로 갑니다.** Thalamus는 조직의 정책만 유지합니다. 권한, 그룹, 큐레이터, 서버 간 연합이 그것입니다.

## 개발

변경하기 전에 [CONTRIBUTING.md](CONTRIBUTING.md)를, 결함을 보고하려면 [SECURITY.md](SECURITY.md)를 읽고, [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)도 읽어 주십시오.

```sh
cd gateway && go test -race ./...           # coverage 100 % outside the two main functions
python3 -I mutate.py                         # one hand mutation per rule, each must fail a test
cd ../extension && npm install && node --test
cd .. && sh tests/e2e_gateway.sh             # a real CortexDB server and the gateway
sh tests/e2e_extension.sh                    # the extension in Chromium
```

각 테스트는 해당 코드보다 먼저 작성되며, 코드 없이는 실패합니다. 전체 절차는 `docs/pipeline_pr.md`에 있습니다. 프로젝트 문서 대부분은 프랑스어로 되어 있습니다.

## 라이선스

MIT. CortexDB도 MIT 라이선스를 따릅니다.

---

이 파일은 [README.md](https://github.com/lfoletti/thalamus/blob/main/README.md)의 번역입니다. 내용이 다를 경우 영어판이 우선합니다. issue는 영어로 작성해 주세요.
