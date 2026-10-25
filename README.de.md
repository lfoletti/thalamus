<p align="center"><img src="extension/icons/icon128.png" alt="Thalamus-Logo: zwei helle Lappen auf einem indigofarbenen Quadrat" width="96"></p>

# Thalamus

[English](https://github.com/lfoletti/thalamus/blob/main/README.md) | [简体中文](https://github.com/lfoletti/thalamus/blob/main/README.zh-CN.md) | [日本語](https://github.com/lfoletti/thalamus/blob/main/README.ja.md) | [한국어](https://github.com/lfoletti/thalamus/blob/main/README.ko.md) | Deutsch

[![CI](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/lfoletti/thalamus/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/lfoletti/thalamus)](https://github.com/lfoletti/thalamus/releases) [![Licence: MIT](https://img.shields.io/github/license/lfoletti/thalamus)](LICENSE)

Thalamus bringt das Wissen einer Organisation zu jedem Benutzer und zu seinem Coding-Agenten, im Rahmen der Rechte jeder Person. Es baut auf [CortexDB](https://github.com/liliang-cn/cortexdb) auf.

> **In Arbeit.** Die Schnittstellen ändern sich zwischen den Versionen.

## Was es tut

Ein Team bewahrt sein Wissen in mehreren CortexDB-Gehirnen auf: eines für das Team, eines für einen Kunden, eines für die ganze Organisation. Jeder Benutzer darf einige davon lesen und in weniger davon schreiben. Thalamus ist das Gateway zwischen den Benutzern und diesen Gehirnen.

Der Assistent des Benutzers erreicht die Gehirne über das Gateway, einen MCP-Server. Wenn der Assistent mit `th_recall` eine Frage stellt, durchsucht das Gateway jedes Gehirn, das dieser Benutzer lesen darf, mit dem eigenen Schlüssel des Benutzers, und führt die Ergebnisse zusammen. Jeder Auszug kommt mit seiner Quelle: dem Gehirn, dem Dokument und dem Autor. Mit `th_get` liest der Assistent das vollständige Dokument hinter einem Auszug.

Ein Benutzer kann auch ein Dokument vorschlagen, aus Claude Code oder von einer Seite in Chrome. Das Dokument wartet auf einen Kurator des Gehirns, der es annimmt oder ablehnt.

## Wozu es dient

### 1. Governance

Thalamus ist eine dünne Governance-Schicht. Es erlaubt einer Organisation, ihr Wissen ruhig und effizient zu teilen, und behält jeden Zugriff unter Kontrolle. Die Verwaltung bleibt einfach: Die Rechte stehen in einer Datei, die Schlüssel folgen den Rechten, und eine Änderung gilt ohne Neustart.

- **Zugang ohne Anfragen.** Ein Benutzer meldet sich mit dem Konto der Organisation an (OpenID Connect). Das Gateway stellt die Schlüssel aus, die die Rechte erlauben. Um einer Person Zugang zu den Gehirnen eines Teams zu geben, fügen Sie das Login der Gruppe des Teams in `rights.yaml` hinzu. Die Änderung gilt beim nächsten Aufruf, ohne Neustart.
- **Ein Schlüssel pro Person und pro Gehirn.** CortexDB prüft den Schlüssel bei jedem Aufruf, daher hängt der Geltungsbereich eines Schlüssels nicht vom Gateway ab. Um einer Person den Zugang zu entziehen, entfernt das Gateway nur die Schlüssel dieser Person.
- **Wissen in einem Schritt hinzugefügt.** Ein Benutzer schlägt ein Dokument aus Claude Code oder von einer Seite in Chrome vor. Ein Kurator des Gehirns nimmt es an oder lehnt es ab, mit denselben Werkzeugen.
- **Einfache Administration.** `th-gateway key list` vergleicht die Schlüssel mit den Rechten, und `key sync` korrigiert die Differenz. `th-gateway audit` gibt eine Zusammenfassung der Lesezugriffe, der Schreibzugriffe, der Ablehnungen und der Anmeldungen.

### 2. Compliance und Souveränität

Thalamus ist darauf ausgelegt, die Einhaltung der DSGVO, des Schweizer Bundesgesetzes über den Datenschutz und, bei einer Hochrisiko-Verwendung, der KI-Verordnung der EU (AI Act) zu unterstützen. Ein Werkzeug macht eine Organisation nicht konform. Aber Thalamus liefert die Kontrollen, die diese Regelwerke verlangen:

| Kontrolle | In Thalamus | Artikel |
| :- | :- | :- |
| Datenminimierung | Die Lesezugriffe werden pro Gehirn und pro Tag gezählt, ohne Namen. Nur ein Gehirn mit `audit_reads: true` hat namentliche Lesezugriffe. | DSGVO 5(1)(c), 25(2). DSG 6(2), 7(3). |
| Speicherbegrenzung | `audit: retention_days` begrenzt, wie lange die Audit-Zeilen bleiben. | DSGVO 5(1)(e). DSG 6(4). |
| Sicherheit und Nachvollziehbarkeit | Persönliche Schlüssel, von der Engine geprüft. Die Schreibzugriffe, die Prüfungen, die Anmeldungen und die abgelehnten Aufrufe haben eine Zeile mit dem Namen der Person. | DSGVO 5(1)(f), 5(2), 32. DSG 8. |
| Speicherort der Daten | Die Organisation betreibt das Gateway und die CortexDB-Server selbst. Die lokalen Gehirne bleiben auf dem Rechner des Benutzers und gehen nicht über das Gateway. | DSGVO 44. DSG 16. |
| Transparenz | Jeder Auszug nennt sein Gehirn, sein Dokument und seinen Autor. | DSGVO 5(1)(a). |
| **Nicht abgedeckt:** die Übermittlung an den Modellanbieter | Der Hook fügt Passagen in den Prompt ein, und der Prompt geht an den Modellanbieter des Assistenten. Thalamus kontrolliert diese Übermittlung nicht. Die Organisation deckt sie selbst ab: ein Auftragsverarbeitungsvertrag mit dem Anbieter, Garantien für die Übermittlung, wenn der Anbieter ausserhalb der Schweiz oder der EU ist, und die Einstellung der Aufbewahrung beim Anbieter. | DSGVO 28, 44 bis 46. DSG 9, 16. |

Ein Gehirn mit `audit_reads: true` führt ein namentliches Lese-Protokoll, und dieses Protokoll ist ein Personendatum: Die Organisation informiert die Benutzer, beantwortet ihre Auskunftsbegehren und legt seine Aufbewahrung fest (DSGVO 13, 15. DSG 19, 25).

**Bei einer Hochrisiko-Verwendung.** Die unten genannten Artikel der KI-Verordnung gelten nur für Hochrisiko-KI-Systeme. Ein Abruf für Coding-Agenten ist im Allgemeinen keine Hochrisiko-Verwendung. Für einen Betreiber eines Hochrisiko-Systems liefert Thalamus die passende Kontrolle:

| Kontrolle | In Thalamus | Artikel |
| :- | :- | :- |
| Aufzeichnungspflichten | Die Schreibzugriffe, die Prüfungen, die Anmeldungen und die abgelehnten Aufrufe haben eine Zeile mit dem Namen der Person. | KI-Verordnung 12. |
| Menschliche Aufsicht | Ein vorgeschlagenes Dokument wartet auf einen Kurator, bevor der Abruf es ausliefert. | KI-Verordnung 14, 26(2). |

DSGVO: Verordnung (EU) 2016/679. DSG: das revidierte Schweizer Bundesgesetz über den Datenschutz (SR 235.1). KI-Verordnung (AI Act): Verordnung (EU) 2024/1689.

### 3. Lokale und geteilte Gehirne in einem Abruf

- **Gehirne in passender Grösse.** Ein Gehirn ist ein ganzer CortexDB-Server oder eine Collection eines geteilten Servers. Kleine Teams können sich einen Server teilen. Ein sensibles Gehirn kann einen eigenen Server haben.
- **Föderation.** Eine Frage erreicht jedes Gehirn, das der Benutzer lesen darf. Das Gateway führt die Ergebnisse zusammen, mit einer Gewichtung pro Gehirn. Der Benutzer wählt nicht, wo gesucht wird.
- **Lokal und geteilt zusammen.** In Claude Code fragt ein Hook bei jedem Prompt gleichzeitig die geteilten Gehirne und die lokalen Gehirne des Rechners ab. Jede Person wählt die Gehirne des automatischen Abrufs mit `/th-recall` (siehe [Lokale und globale Gehirne kombinieren](#combining-local-and-global-brains)).

### 4. Überprüfbare Genauigkeit

- **Passagen mit ihrer Quelle.** Der Abruf liefert die Passagen der Dokumente, nicht kurze Auszüge. Mit `th_get` liest der Agent das vollständige Dokument.
- **Eine Pyramide von Synthesen.** Ein Gehirn kann über seinen Dokumenten eine Collection von Synthesen haben. `th_recall` liefert zuerst die Synthesen, und `th_descend` öffnet die Dokumente, die eine Synthese zitiert.
- **Gemessen.** In einem Forschungspilot stieg der Anteil der vom automatischen Abruf gefundenen Fakten von 51,7 % auf 81,3 % mit den Passagen und den Synthesen, und auf 85,1 % mit einer Variante der Frage auf Englisch.
- **Entitäten über Gehirne hinweg.** `th_entity` findet eine Entität in jedem Gehirn, das der Benutzer lesen darf. Es führt die Nachbarn nach Knoten-ID zusammen und nennt die Gehirne, die jeden davon liefern.
- **Als Nächstes.** Eine Prüfung jedes Satzes einer Synthese gegen ihre Quellen, mit dem Werkzeug von CortexDB v2.139.0.

### 5. Mehrsprachig

Ein Benutzer fragt in seiner Sprache, und das Dokument, das antwortet, ist oft in einer anderen. `th_recall` nimmt die Frage in weiteren Sprachen entgegen (`alternate_queries`) und sucht jede einzeln.

- Der Schweizer Benchmark (`tests/eval_ch/`) hat 40 Fragen zur Bundesverfassung, auf Deutsch, Französisch, Italienisch und Rätoromanisch. Jede Frage verlangt einen Artikel in einer anderen Sprache.
- Mit `multilingual-e5-large` und den Varianten sind 97 % der erwarteten Artikel unter den ersten 10 Ergebnissen, und 91 % für Rätoromanisch. Ohne Encoder sind es nur 7 %.

Für Coding-Agenten bringt Thalamus das Wissen in die Sitzung: über MCP für jeden Assistenten und über einen automatischen Abruf bei jedem Prompt in Claude Code.

## Begriffe

| Begriff | Bedeutung |
| :- | :- |
| Gehirn | Ein CortexDB-Server oder eine Collection eines geteilten Servers. |
| Rechte | Die Datei `rights.yaml`. Sie deklariert die Gehirne, die Gruppen und wer jedes Gehirn liest, beschreibt oder kuratiert. |
| Schlüssel | Ein CortexDB-API-Schlüssel für eine Person auf einem Gehirn. Das Gateway schreibt ihn in die Schlüsseldatei des Servers, beschränkt auf die Collection des Gehirns, nur lesend oder lesend und schreibend. |
| Kurator | Eine Person, die die einem Gehirn vorgeschlagenen Dokumente annimmt oder ablehnt. |

## Wie es funktioniert

<p align="center"><img src="docs/img/architecture.svg" alt="Das Claude-Code-Plugin, die Chrome-Erweiterung und th-gateway login rufen das Thalamus-Gateway auf. Das Gateway liest rights.yaml, schreibt audit.jsonl und ruft die CortexDB-Server mit dem Schlüssel des Aufrufers auf: ein Schlüssel pro Person und pro Gehirn." width="760"></p>

- Das Gateway liest `rights.yaml` und stellt die Schlüssel aus. Es liest die Datei erneut, wenn sie sich ändert, daher gilt ein entferntes Recht beim nächsten Aufruf.
- Jeder Aufruf geht mit dem Schlüssel des Aufrufers an CortexDB. CortexDB prüft den Schlüssel bei jedem Aufruf, daher hängt der Geltungsbereich eines Schlüssels nicht vom Gateway ab.
- Das Gateway führt ein Audit in `audit.jsonl`, für Sicherheit und Nachvollziehbarkeit. Die Schreibzugriffe, die Prüfungen, die Anmeldungen und die abgelehnten Aufrufe haben eine Zeile mit dem Namen der Person. Die Lesezugriffe werden pro Gehirn und pro Tag gezählt, ohne Namen. Nur ein Gehirn mit `audit_reads: true` in `rights.yaml` hat namentliche Lesezugriffe. `audit: retention_days` begrenzt, wie lange die Zeilen bleiben.
- Ein Benutzer meldet sich beim Identitätsanbieter der Organisation an (OpenID Connect). Das Gateway stellt dann die Schlüssel aus, die die Rechte erlauben.

## <a id="combining-local-and-global-brains"></a>Lokale und globale Gehirne kombinieren

Das Thalamus-Plugin für Claude Code fügt den Werkzeugen des Gateways einen Hook hinzu. Bei jedem Prompt fragt der Hook gleichzeitig die geteilten Gehirne und die lokalen Gehirne des Rechners ab. Er gibt dem Agenten die besten Auszüge, bevor der Agent antwortet. Der Benutzer fragt nicht danach.

<p align="center"><img src="docs/img/recall-hook.svg" alt="Der Benutzer tippt einen Prompt. Der Thalamus-Hook sendet ihn an th_recall auf dem Thalamus-Gateway, mit dem Schlüssel des Benutzers, und das Gateway durchsucht die Server-Gehirne. Gleichzeitig durchsucht der Hook die lokalen Gehirne auf dem Rechner. Der Hook fügt die besten Auszüge mit ihrer Quelle dem Prompt hinzu. Dann antwortet der Agent." width="760"></p>

- Der Hook (`th-hook`, beim Ereignis `UserPromptSubmit`) sendet den Prompt mit dem Schlüssel des Benutzers an `th_recall`. Er fügt dem Prompt die besten Passagen der Dokumente mit ihrer Quelle hinzu, in höchstens 9000 Zeichen: 10 Passagen, nach 4 Passagen von Synthesen, wenn ein Gehirn eine Pyramide von Synthesen hat.
- Er fragt nicht bei einem Prompt mit weniger als 3 Wörtern oder bei einem `/`-Befehl.
- Er hält nie einen Prompt auf. Wenn das Gateway nicht antwortet oder der Schlüssel abgelehnt wird, läuft der Prompt ohne Auszüge weiter.
- Der Agent kann mit `th_get` das vollständige Dokument eines Auszugs öffnen.

- Für die lokalen Gehirne führt der Hook den nativen Abruf von CortexDB aus, einmal pro Gehirn. Standardmässig ist das einzige lokale Gehirn das Gehirn des Projekts. Eine Einstellung fügt weitere lokale Gehirne hinzu, zum Beispiel das Gehirn im Home-Verzeichnis oder das Gehirn eines anderen Projekts.
- Die lokalen Gehirne bleiben auf dem Rechner. Sie gehen nicht über das Gateway und stehen nicht im Audit.
- Der Block enthält zuerst die geteilten Auszüge, dann einen Abschnitt pro lokalem Gehirn. Die lokalen Abschnitte belegen höchstens 3000 Zeichen.
- Jede Person wählt die Gehirne des automatischen Abrufs, lokal und geteilt, mit `/th-recall` (`on`, `off`, `only`). Die Wahl ändert die Rechte nicht.

Andere Assistenten nutzen dieselben Werkzeuge über MCP. Der automatische Abruf braucht einen Hook im Assistenten. Thalamus hat einen für Claude Code.

## Installation

Jedes Release enthält diese Dateien:

- `thalamus_<version>_<os>_<arch>.tar.gz`: die Binärdateien `th-gateway` und `th-hook`, für macOS und Linux, amd64 und arm64.
- `thalamus-plugin_<version>_<os>_<arch>.tar.gz`: das Claude-Code-Plugin, mit `th-hook` in `bin/`. Starten Sie Claude Code mit `claude --plugin-dir thalamus-plugin`.
- `thalamus-extension_<version>.zip`: die Chrome-Erweiterung. Entpacken Sie sie, öffnen Sie `chrome://extensions`, schalten Sie den Entwicklermodus ein und klicken Sie dann auf **Entpackte Erweiterung laden**.
- `checksums.txt`: der SHA-256 jeder Datei.

`th-gateway version` und `th-hook version` geben die Release einer Binärdatei aus, und `GET /v1/version` auf einem Gateway antwortet ohne Schlüssel: Ein Deployment wird gegen die erwartete Release geprüft.

Mit Go installieren Sie die Binärdateien aus dem Quellcode:

```sh
go install github.com/lfoletti/thalamus/gateway/cmd/th-gateway@latest
go install github.com/lfoletti/thalamus/gateway/cmd/th-hook@latest
```

Das Gateway gibt es auch als Image. Legen Sie `rights.yaml` und die Schlüsseldateien in ein Verzeichnis und binden Sie es dann unter `/etc/thalamus` ein:

```sh
docker run -p 8787:8787 -v /srv/thalamus:/etc/thalamus ghcr.io/lfoletti/thalamus:latest
```

Die Audit-Datei liegt in `/var/lib/thalamus`. Binden Sie dort ein Volume ein, um sie zu behalten. Das Gateway braucht für jedes Gehirn einen CortexDB-gRPC-Server (siehe [Ausprobieren](#try-it)).

## <a id="try-it"></a>Ausprobieren

Die Schritte starten einen CortexDB-Server mit zwei Gehirnen, dann das Gateway, dann das Plugin. Führen Sie alle Schritte im Wurzelverzeichnis dieses Repositorys aus.

1. Klonen Sie CortexDB nach `cortexdb/`, neben `gateway/`, und bauen Sie dann die Binärdateien.

   ```sh
   git clone https://github.com/liliang-cn/cortexdb.git && (cd cortexdb && git checkout v2.139.0)
   mkdir -p bin
   (cd cortexdb && go build -o ../bin/cortexdb-grpc ./cmd/cortexdb-grpc)
   (cd gateway && go build -o ../bin/th-gateway ./cmd/th-gateway && go build -o ../plugin/bin/th-hook ./cmd/th-hook)
   ```

2. Schreiben Sie die Rechte und eine leere Schlüsseldatei.

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

   Die Schlüsseldatei liegt unter dem Pfad, den `keys_file` angibt. Ein relativer Pfad beginnt in dem Verzeichnis, in dem `th-gateway` läuft. Das Gateway schreibt die Schlüssel in diese Datei, und der Server liest sie mit `-keys`.

3. Stellen Sie die Schlüssel aus und starten Sie dann den Server. Der Server liest seine Schlüsseldatei beim Start.

   ```sh
   bin/th-gateway key sync rights.yaml
   bin/cortexdb-grpc -db brain.db -addr 127.0.0.1:47998 -keys keys.json &
   ```

4. Starten Sie das Gateway.

   ```sh
   bin/th-gateway serve rights.yaml --listen 127.0.0.1:8787 &
   ```

5. Starten Sie Claude Code mit dem Plugin. Verwenden Sie eines der Geheimnisse, die `key sync` ausgegeben hat.

   ```sh
   export TH_GATEWAY=http://127.0.0.1:8787 TH_KEY=<secret>
   claude --plugin-dir ./plugin
   ```

Mit einem `identity`-Block in `rights.yaml` ersetzt `th-gateway login http://127.0.0.1:8787` den Export von Schritt 5. Die Einzelheiten stehen in `gateway/README.md`, `plugin/README.md` und `extension/README.md`.

## Roadmap

**Erledigt** (25. Oktober 2026, Release 0.4.0)

- Rechte, Schlüssel und ein MCP-Endpunkt pro Gehirn.
- Föderierter Abruf über die Gehirne, die eine Person lesen darf, mit der Quelle jedes Auszugs.
- Vorgeschlagene Dokumente und ihre Prüfung durch einen Kurator.
- REST-Endpunkte für Clients ohne MCP.
- Anmeldung mit OpenID Connect, von der Kommandozeile und aus der Erweiterung.
- Das Claude-Code-Plugin: Abruf-Hook, Werkzeuge und der Befehl `/th-add`.
- Die Chrome-Erweiterung: eine Seite oder eine Auswahl vorschlagen. Sie behält den Hauptinhalt der Seite.
- Ein verhältnismässiges Audit: namentliche Schreibzugriffe, Prüfungen, Anmeldungen und Ablehnungen, ohne Namen gezählte Lesezugriffe, namentliche Lesezugriffe nur auf den Gehirnen, die es verlangen, und eine Aufbewahrungsfrist.
- Lokale und geteilte Gehirne in einem Abruf: Der Hook fügt die lokalen Gehirne des Rechners hinzu, und jede Person wählt die Gehirne des automatischen Abrufs.
- Abruf auf den Passagen der Dokumente, nicht auf kurzen Auszügen.
- Eine Pyramide von Synthesen pro Gehirn: `th_recall` liefert zuerst die Synthesen, und `th_descend` öffnet die Dokumente, die eine Synthese zitiert.
- Entitäten über Gehirne hinweg: `th_entity` fragt jedes Gehirn nach einer Entität und ihren Nachbarn, mit dem Schlüssel des Benutzers für dieses Gehirn, und führt die Knoten nach ID zusammen.
- Eine Audit-Zeile für jeden Schlüssel, den ein Gehirn während `th_recall` oder `th_entity` ablehnt.
- Fragen in mehreren Sprachen: `th_recall` nimmt die Frage in weiteren Sprachen entgegen (`alternate_queries`) und sucht jede einzeln. Ein Schweizer Benchmark misst den Abruf von einer Sprache in eine andere (Französisch, Deutsch, Italienisch, Rätoromanisch) in `tests/eval_ch/`.
- Eine Frist für den Abruf: `th_recall` nimmt `deadline_ms` entgegen und führt die Antworten der Gehirne zusammen, die bereit sind. Der Hook nutzt sie, sodass ein langsames Gehirn den Abruf der anderen nicht mehr aufhält.
- Die Gehirne in Claude Code: `/brains` öffnet einen Dialog der Gehirne, der zeigt, ob jedes antwortet, und Enter schaltet ein Gehirn für den automatischen Abruf ein oder aus. `/th-recall` ohne Argumente fragt mit einer Checkliste. Die `CLAUDE.md` des Projekts nennt das Gehirn, in das `th_add` schreibt.
- Releases: Archive von `th-gateway`, `th-hook`, dem Plugin und der Erweiterung für macOS und Linux, ein Docker-Image `ghcr.io/lfoletti/thalamus` und `go install`. Siehe die [Releases](https://github.com/lfoletti/thalamus/releases).

**Mit CortexDB**

Thalamus stützt sich für jeden Mechanismus, der allen CortexDB-Benutzern dient, auf CortexDB. Diese Teile gehen als kleine Pull Requests an CortexDB.

- Gemergt: die Allowlist der Werkzeuge für einen beschränkten Schlüssel (#72 bis #74, #78), Inhalts-Hash und Duplikatprüfung (#75), Ablaufdaten (#76), Prüfstatus (#77) und weitere Korrekturen (#61, #67 bis #71, #82, #83).
- Gemergt in v2.139.0, aus der Diskussion #86:
  - Schlüssel und Signaturen: Der Server liest seine Schlüsseldatei erneut, wenn sie sich ändert, sodass ein neuer oder widerrufener Schlüssel ohne Neustart gilt (#87). Jeder Schreibzugriff wird mit dem Schlüssel des Aufrufers signiert (#88). Eine Erinnerung wird nur gelesen, wenn der Geltungsbereich des Aufrufers es erlaubt (#89).
  - Graph-Werkzeuge für einen auf eine Collection beschränkten Schlüssel (#94 bis #98). `th_entity` nutzt sie, um Entitäten über Gehirne hinweg zusammenzuführen, zusätzlich zu den Passagen.
  - Drei Werkzeuge und ein Index: die berührten IDs im Request-Log (#90), `knowledge_get` auf einer Liste von IDs (#91), eine Prüfung eines Satzes gegen seine zitierten Quellen (#92), und zwei redundante Kanten-Indizes entfernt (#93).
  - Suche in europäischen Sprachen: ein Schweizer Evaluationsdatensatz (#99), Akzente in der Abdeckungsuntergrenze (#100), und die Wörter und Läufe alternativer Anfragen (#101, #102).
- Vorgeschlagen: `upsert_entities` erfasst die Collections der Chunks, die es nennt (#105), damit `th_entity` die Entitäten eines ausserhalb von Thalamus gebauten Gehirns liefert.
- Thalamus verwendet CortexDB v2.139.0.
- Vorbereitet: ein Werkzeug, das die Dokumente auflistet, die auf eine Prüfung warten.

**Als Nächstes**

- Entitäten für die über `th_add` gefüllten Gehirne: Das Gateway hat noch keinen Extraktor.
- Ein einfacher Installationsweg für `th-hook` und die Erweiterung im Chrome Web Store.
- Eine kurze Demonstration im README.

## Grundsätze

1. **Die Engine prüft jeden Schlüssel.** Das Gateway entscheidet, wer einen Schlüssel bekommt. CortexDB entscheidet, was jeder Schlüssel tun darf.
2. **Ein Schlüssel pro Person und pro Gehirn.** Das Gateway kann den Zugang einer Person auf einem Gehirn widerrufen.
3. **Jeder Auszug zeigt seine Quelle.** Der Leser kann mit `th_get` das Dokument hinter einem Auszug öffnen.
4. **Allgemeine Mechanismen gehen upstream.** Thalamus behält nur die Policy der Organisation: die Rechte, die Gruppen, die Kuratoren und die Föderation zwischen den Servern.

## Entwicklung

Lesen Sie [CONTRIBUTING.md](CONTRIBUTING.md) vor einer Änderung, [SECURITY.md](SECURITY.md), um einen Fehler zu melden, und [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

```sh
cd gateway && go test -race ./...           # coverage 100 % outside the two main functions
python3 -I mutate.py                         # one hand mutation per rule, each must fail a test
cd ../extension && npm install && node --test
cd .. && sh tests/e2e_gateway.sh             # a real CortexDB server and the gateway
sh tests/e2e_extension.sh                    # the extension in Chromium
```

Jeder Test wird vor seinem Code geschrieben und schlägt ohne ihn fehl. Das vollständige Vorgehen steht in `docs/pipeline_pr.md`. Die meisten Projektdokumente sind auf Französisch.

## Lizenz

MIT. CortexDB steht ebenfalls unter der MIT-Lizenz.

---

Diese Datei ist eine Übersetzung von [README.md](https://github.com/lfoletti/thalamus/blob/main/README.md). Bei Abweichungen gilt die englische Fassung. Bitte schreiben Sie Issues auf Englisch.
