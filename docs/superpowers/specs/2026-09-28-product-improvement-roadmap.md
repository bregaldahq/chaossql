# ChaosSQL: análise de melhorias e roadmap

Data: 2026-09-28. Base: `main` em `03336b7` (v1.6.0).

Método: leitura do código, das seções "Gotchas" de todas as skills (fatos
verificados), do `site/COPY.md`, do CHANGELOG, e execução real da CLI
compilada a partir deste commit. Tudo marcado como **[verificado]** foi
reproduzido durante esta análise.

---

## Resumo executivo

O motor é sólido e a engenharia está acima da média: agendamento
determinístico, ddmin com auditoria de 1-minimalidade, classificação Adya,
replay executável, Cloud com privacidade por projeção, harness de agentes.
O site novo está bom.

O gargalo agora **não é falta de feature. É que a primeira experiência não
entrega a promessa**, e o produto ainda não chega nas mãos das pessoas:

1. O quickstart do README e o `chaossql demo` **não mostram bug nenhum**.
2. Uma violação **não falha o CI** (exit 0), então o "guardião de PR" não guarda.
3. `pip install chaossql` e `npm install @chaossql/test` dão **404**, e a
   release não tem **nenhum binário**.
4. O playground roda sobre o driver Mock (tudo zero), então não demonstra
   isolamento de verdade.
5. 0 estrelas, 0 forks: falta distribuição e uma história de "bug real
   encontrado".

Recomendação: fazer uma **v1.7 "primeiros 5 minutos"** focada nos itens P0
antes de qualquer feature nova. São mudanças pequenas com impacto enorme na
ativação.

---

## P0: a primeira experiência (fazer já)

### P0.1 O quickstart e o demo passam sem encontrar o bug [verificado]

```
chaossql run examples/banking_lost_update/chaos.yaml   → ✔ ALL INVARIANTS SATISFIED
chaossql demo banking_lost_update                      → ✔ ALL INVARIANTS SATISFIED
chaossql init x && chaossql run x/chaos.yaml           → ✔ ALL INVARIANTS SATISFIED
```

Causa: no SQLite o nível padrão é `SERIALIZABLE` com uma conexão presa durante
a transação, então tudo roda em série. Nenhum `examples/*/chaos.yaml` define
`isolation`. Com `isolation: READ_UNCOMMITTED` o mesmo cenário dá
`P4_LOST_UPDATE` reduzido a 2 operações.

A pessoa segue o README, vê "tudo certo" e fecha a aba. É o maior vazamento
do funil.

Proposta:
- Adicionar `isolation: READ_UNCOMMITTED` nos exemplos SQLite que devem
  falhar, e fazer o `demo` usar o nível que expõe a anomalia.
- Adicionar as flags `--isolation`, `--driver` e `--dsn` no `run`. Hoje nenhuma
  delas existe, então testar o mesmo cenário em outro nível exige editar o YAML.
- Quando um run passa no SQLite em `SERIALIZABLE`, imprimir uma dica:
  "SQLite executou as transações em série; para caçar anomalias use
  `--isolation READ_UNCOMMITTED` ou PostgreSQL/MySQL".
- Fazer o `init` gerar um cenário que **falha** e depois mostrar como
  corrigir. O "aha" é ver o bug, não ver o verde.
- Teste de regressão: um teste no CI que roda o comando exato do quickstart e
  exige `violation`, para isso não voltar.

**Status (2026-09-28): implementado.** Quickstart, `demo` (agora com os
exemplos embutidos no binário, pois antes falhava após `go install`) e `init`
mostram a lost update; flags `--driver/--dsn/--isolation`; `diff`, `matrix` e
`swarm` caem para o nível padrão do banco quando ele não suporta o nível do
spec. Achados novos ao validar em PostgreSQL 16 real:
- Os schemas dos exemplos usam sintaxe SQLite (`AUTOINCREMENT`): 7 de 10
  falham no reset em PostgreSQL. Por isso `matrix --driver postgres` marca
  tudo como "PREVENTED" (erro conta como seguro, item P1).
- No PostgreSQL, o scaffold do `init` acha a violação, mas o classificador
  rotula como `A5A_READ_SKEW` em vez de P4 (reforça o item 1 da tabela P1).
- A tabela de flags da documentação do site (`docs.json`) lista flags que não
  existem (`--duration`, `--jitter-min`) e ainda não tem as novas.

### P0.2 Violação não falha o build [verificado]

`chaossql run` retorna exit 0 em `violation`. O GitHub Action fica verde. O
bloqueio só existe via JUnit ou via `is-regression`, e este último exige Cloud.

Para uma ferramenta vendida como "guarda de PR", isso é o comportamento
errado por padrão.

Proposta:
- `--fail-on=violation|regression|never` no `run` (padrão `violation` numa
  versão major, ou já agora no Action com input `fail-on-violation: true`).
- Códigos de saída distintos e documentados: 0 passou, 1 violação, 2 erro de
  execução, 3 inconclusivo.
- O mesmo para `diff`/`swarm`, que retornam 0 mesmo com divergência.
- Atualizar `site/COPY.md` ("A violation does not fail the build") e as
  skills `chaossql-cli-run-pipeline` e `chaossql-github-action`.

**Status (2026-09-28): implementado.** Decisão: reprovar por padrão (breaking
change registrado no CHANGELOG). Códigos 0/1/2, `--fail-on` no `run` e no
Action, `--fail-on-divergence` opt-in no `diff`/`swarm` (o `swarm.yml`
compara motores onde a divergência é esperada). O `concurrency-ci.yml` virou
teste do próprio gate: ele exige que o Action falhe no exemplo com bug.

### P0.3 Distribuição: não dá para instalar fora do Go [verificado]

- Release v1.6.0: **0 assets**. Só existe `go install`, o que exige Go 1.25.
- PyPI `chaossql` e npm `@chaossql/test`: **404**, mas o README manda instalar.
- As versões dos SDKs (1.4.0) estão atrás do Go (1.6.0) e ficam em 3 lugares
  no Python.
- O Action compila do fonte a cada job (setup-go + build): lento e frágil.

Proposta:
- GoReleaser: binários linux/darwin/windows (amd64/arm64), checksums, SBOM,
  Homebrew tap, imagem `ghcr.io/bregaldahq/chaossql`, script `install.sh`.
- Action baixa o binário da release (com cache), com fallback para build.
- Publicar os SDKs no PyPI e no npm com trusted publishing no mesmo workflow
  de release, e fazer o SDK baixar o binário se não achar no PATH.
- Checagem no `make check-harness` de que todas as versões batem com
  `internal/version`.

### P0.4 O playground não mostra isolamento de verdade

O WASM usa o driver Mock: toda query retorna `0`. O resultado depende só de
como a assertion se comporta com zeros. O site diz "See the bug happen", mas
o que o playground mostra é outra coisa.

Curto prazo: no playground, tocar os **traces gravados** reais (já existem
em `site/src/data/traces/`) com rótulo claro, e deixar o editor livre como
"modo avançado, simulado".

Médio prazo, e também uma feature de produto: ver **F1 (driver simulador)**
abaixo.

### P0.5 README desalinhado com o próprio guia de copy

O `site/COPY.md` proíbe números inventados, mas o README abre com um bloco
fictício ("$1,950.00", "seed 184729", "< 200ms"). Ele também manda
`pip install` e `npm install` (404), usa `go-version: '1.23'` no exemplo SARIF
(o go.mod exige 1.25) e usa `|| true` para esconder o exit code. Proposta:
aplicar ao README as mesmas regras do site, colocar o GIF real no topo (já
está) e trocar o bloco fictício pelo output real.

---

## P1: confiança no resultado

Uma ferramenta de correção perde o usuário no primeiro falso positivo ou
rótulo errado. Estes itens são pequenos e cada um remove uma dúvida.

| # | Problema (fonte: skill) | Proposta |
| :-: | :--- | :--- |
| 1 | O rótulo da anomalia depende da ordem (possivelmente aleatória) dos ciclos. HTML, UI, SARIF e Mermaid têm cada um seu próprio loop de prioridade, então os rótulos podem divergir. `read_skew_financial_audit` alterna entre P4 e A5A. | Uma função única `ClassifyTrace` com ordenação determinística dos ciclos, usada por todos os reporters. Teste de estabilidade do rótulo por exemplo. |
| 2 | O SARIF emite finding para run **aprovado** (se houver ciclos ou se o nome do cenário contiver "deadlock"). | Emitir só para `violation`. Ciclo sem violação vira nota, não erro. |
| 3 | `ticket_booking_anti_dependency` falha até em histórico serial: é um falso positivo num exemplo canônico. | Corrigir o invariante ou o cenário e adicionar teste "serial deve passar" para todos os exemplos. |
| 4 | Campos aceitos e ignorados em silêncio: `temporal_invariants` (nunca roda no `run`), `weight` (ignorado), `faults.disconnect_probability` (sem efeito). | Implementar ou fazer o `validate` e o parser rejeitarem com erro claro. Spec que mente é pior que spec que falta. |
| 5 | O `validate` compila a assertion sem ambiente, então não pega coluna inexistente nem tipo errado (`total == 2000` com string vira `inconclusive`). | Um modo `validate --db` que roda schema+seed e checa a assertion contra as colunas reais. |
| 6 | Só o primeiro invariante que falha é avaliado. | Avaliar todos e reportar a lista. |
| 7 | Geradores com argumento inválido passam em silêncio (`$random_int(5)` vai cru para o SQL). `{...}` come JSON literal. Aritmética só aceita um `+`/`-`. | Erros de validação no parse. Escape para chaves (`{{`). Usar o `expr` já presente para as expressões. |
| 8 | MySQL: valores chegam como `[]byte`, então as captures viram `[49 48 48 48]` e os invariantes viram string. O MySQL é anunciado, mas as captures estão quebradas nele. | Normalizar `[]byte` para número/string no driver MySQL, com testes de integração. |
| 9 | O `Reset` no PostgreSQL destrói o schema `public` inteiro e no MySQL dropa todas as tabelas do banco. | Rodar em schema dedicado (`chaossql_<run>`), ou exigir `--allow-destructive-reset` / nome de banco com `test`. Isso também habilita runs paralelos. |
| 10 | O DSN não aceita variáveis de ambiente no `run`, então a senha fica commitada no YAML e vai parar no artefato de replay. | `${ENV}` no DSN (o swarm já resolve DSN via env) e redação do DSN no artefato. |
| 11 | SDK Python lê `failing_invariant["observed"]`, uma chave que o engine nunca envia. Defaults diferentes entre SDKs e CLI (workers, iterations, jitter). | Corrigir. Unificar os defaults num único lugar e documentar. |
| 12 | Cloud: erro de infraestrutura conta como regressão. `/v1/health` responde versão fixa `1.0.0`. | Separar `execution_error` de `failed` no veredito. Usar `version.Version`. |
| 13 | Proxy: aresta WW nunca é criada para escritor não commitado, prepared statements re-executados ficam invisíveis, sem TLS, grafo nunca é podado. | Corrigir o bug do WW (é pontual). Documentar os limites no output do proxy. Poda por janela de transações concluídas. |

---

## P2: features para crescer

Em ordem de impacto sobre esforço.

### F1. Driver simulador determinístico (`driver: sim`) [alto impacto]

Um banco em memória em Go puro que **modela** RC, RR, SI e SERIALIZABLE
(MVCC com regras explícitas) para o subconjunto de SQL dos cenários (tabela
de chave primária, SELECT/UPDATE/INSERT/DELETE por chave e por predicado
simples).

Por que vale:
- Roda no WASM: o playground passa a mostrar anomalias **reais** por nível
  de isolamento, no browser, sem servidor. Resolve o P0.4.
- É 100% determinístico, inclusive o intercalamento físico. O replay passa a
  ser bit-a-bit e o ddmin deixa de ser flaky.
- É rápido: milhares de schedules por segundo, bom para pre-commit e CI barato.
- Serve de "oráculo de referência" para o `diff` contra os bancos reais.

Risco: não substitui o banco real. Deve ser posicionado como modelo, e o
`matrix` deve comparar sim contra real.

### F2. Testar o código real da aplicação, não só YAML [alto impacto]

Hoje a pessoa precisa **reescrever** a lógica de transação em templates SQL.
É a maior barreira de adoção para quem tem ORM.

Proposta: um wrapper de `database/sql` (Go) que injeta pontos de
escalonamento determinísticos em cada `Exec`/`Query`/`Commit`. O teste chama
as funções reais da aplicação em N goroutines e o ChaosSQL controla a
intercalação por seed. Depois, equivalentes para Python (eventos do
SQLAlchemy / wrapper DB-API) e Node (`pg`, Prisma middleware).

```go
db := chaostest.Wrap(t, realDB, chaostest.Seed(42))
chaostest.Concurrently(t, db, 2, func(db *sql.DB) { app.Withdraw(db, 1, 100) })
chaostest.Invariant(t, db, "SELECT ...", "balance >= 0")
```

Isso transforma o ChaosSQL de "ferramenta de cenário" em "biblioteca de
teste", que é onde está o hábito diário.

### F3. `chaossql record`: gerar o cenário a partir do tráfego real

O proxy já vê os statements. Gravar as transações de uma sessão (dev ou
staging), agrupar por template (parametrizando literais) e emitir um
`chaos.yaml` pronto, com sugestão de invariantes (somas conservadas,
`>= 0`, unicidade). Isso reduz o tempo até o primeiro cenário útil de horas
para minutos.

### F4. Do bug ao conserto verificado

Depois de classificar, sugerir a correção (já está na tabela do README:
`FOR UPDATE`, update atômico, `UNIQUE`, SSI) e oferecer
`chaossql verify-fix --patch fix.yaml`, que roda **a mesma seed** e mostra
"antes: P4 em 2 ops / depois: 0 violações em N schedules". Fecha o ciclo
"encontrar, reproduzir, corrigir, provar" e é um ótimo material de demo.

### F5. Bancos efêmeros com um flag

`--ephemeral postgres:16` / `mysql:8`: sobe um container (Testcontainers
ou `docker run`), roda e derruba. Resolve o medo do `Reset` destrutivo (P1.9)
e o atrito de configurar DSN. Receitas prontas para GitHub Actions `services:`.

### F6. Cobertura de exploração

Responder "quanto tempo devo fuzzar?". Reportar schedules distintos, pares
de conflito cobertos e profundidade PCT atingida, com a garantia
probabilística do PCT para profundidade d. Útil no Cloud como gráfico por
cenário.

### F7. Ecossistema e DX baratos

- JSON Schema do `chaos.yaml` publicado no SchemaStore: autocomplete e
  validação no VS Code/JetBrains sem extensão.
- Testar e anunciar CockroachDB, YugabyteDB, Neon e Supabase (protocolo PG
  via pgx, custo baixo) e MariaDB.
- Receitas por ORM na documentação: Prisma, Django, Rails, SQLAlchemy, GORM.
- Pre-commit hook e `chaossql run --changed` (roda só os cenários afetados).

---

## P3: Cloud e monetização

Estado atual: planos existem (`billing.go`), mas **não há cobrança**
(nenhuma integração de pagamento) e as feature flags dos planos não são
aplicadas. Hoje o Cloud é, na prática, lista de espera e auditoria.

Sugestão de ordem:
1. **Não investir em billing antes de ter ativação.** Os P0 vêm primeiro. A
   auditoria de $1,490 é a receita realista de curto prazo. Colocar o CTA da
   auditoria no output da CLI quando houver violação em PostgreSQL/MySQL
   (discreto, desligável).
2. **GitHub App** no lugar de token manual: instalação em 1 clique, comentário
   de PR sem configurar `CHAOSSQL_CLOUD_TOKEN` + `GITHUB_TOKEN`.
3. **Fuzzing agendado (nightly)** no Cloud. Está listado em "Not claimed on
   purpose" e é a feature que mais justifica pagar: roda milhares de seeds
   fora do CI do PR e abre issue quando acha algo.
4. Dashboard: histórico por cenário, detecção de flaky (mesma seed, resultados
   diferentes), tendência de cobertura (F6).
5. Paginação nas listas (fixo em 50 hoje) e checkout (Stripe) quando houver
   demanda.
6. Escala: SQLite com 1 réplica atende por bastante tempo. Não mexer agora.

---

## Crescimento e distribuição

0 estrelas e 0 forks, com o produto neste nível, indicam um problema de
**distribuição**, não de produto.

1. **Uma história de bug real.** Rodar o ChaosSQL contra 3 a 5 projetos open
   source populares com lógica de saldo ou estoque (e-commerce, fintech,
   reservas), achar um lost update ou write skew de verdade, reportar
   responsavelmente e escrever o post. Um bug real vale mais que qualquer
   landing.
2. **Página de referência "isolation levels comparados"** com a matriz
   medida de verdade em PostgreSQL 16, MySQL 8 e SQLite, regenerada no CI.
   É um conteúdo que atrai buscas orgânicas ("postgres repeatable read write
   skew"). Isso exige que o `matrix` rode contra bancos reais no CI e publique
   o JSON que o site consome.
3. Lançamento coordenado depois dos P0: Show HN, r/golang, r/PostgreSQL,
   Lobsters, dev.to. Em PT-BR: TabNews e comunidades Go/Postgres Brasil.
4. Listas: awesome-go, awesome-postgres, awesome-testing. Palestra curta
   ("Encontrando lost updates em 2 minutos") em GopherCon BR / PGConf.
5. Medir: o `tools/site_funnel.mjs` já existe. Adicionar um evento
   "install copiado" e "primeira violação vista" (opt-in na CLI, ou via
   comando `demo` que abre o site) para medir ativação de verdade.

---

## Saúde de engenharia (rápidos)

- `swarm.yml` fixa Go 1.23 enquanto o `go.mod` exige 1.25.
- `wasm_exec.js` commitado é anterior ao toolchain 1.25.
- `--export-repro`/`--export-mermaid` ignoram caminho e sempre escrevem no CWD.
- `bench` com cenário ignora o workload do cenário (sempre `accounts`).
- `--json` embute HTML/OTLP/repro inteiros: oferecer `--json=compact`.
- O SQLite `Reset` não remove views/triggers de schemas anteriores.

---

## Roadmap sugerido

| Versão | Tema | Itens |
| :--- | :--- | :--- |
| **v1.7** (1 a 2 semanas) | Primeiros 5 minutos | P0.1, P0.2, P0.3, P0.5; P1 itens 2, 3, 11, 12 |
| **v1.8** | Confiança | P1 itens 1, 4 a 10; playground com traces reais (P0.4 curto) |
| **v1.9** | Simulador | F1 (driver `sim` + playground real) e F7 (JSON Schema, Cockroach/Neon) |
| **v2.0** | Código real | F2 (wrapper `database/sql`), F4 (verify-fix), F5 (efêmero); exit codes novos por padrão |
| Em paralelo | Distribuição | Caça a bug real em OSS, página da matriz, lançamento após v1.7 |
| Depois | Cloud pago | GitHub App, nightly fuzzing, F3 (record), billing |

Critério de sucesso da v1.7: uma pessoa sem Go instalado vai de zero até ver
uma lost update reduzida a 2 operações, e um PR vermelho por causa dela, em
menos de 5 minutos.
