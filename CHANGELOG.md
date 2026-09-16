# Changelog - liturgical-calendar-engine

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/), e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [2.4.0] - 2026-09-16

### 📑 Feature: Endpoint de Exportação do Calendário Litúrgico e Mariano (`.xls` e `.html`)
- **Novo Endpoint de Exportação (`GET /api/v1/calendar/export` e `GET /calendar/export`)**:
  - Parâmetros suportados: `year`, `month`, `calendar` (1962 ou 1954), `lang` (padrão: pt-br), `format` (padrão: `xls`, opção `html`), `include_brazilian` (padrão: true).
  - Gera tabela padronizada em 4 colunas em conformidade estrita com o modelo tradicional do Salve Maria:
    1. **Dia**: Dia do mês e dia da semana em português (ex: `1 de Outubro\nQuinta`).
    2. **Calendário Litúrgico**: Festa principal, comemorações formatadas e grau/classe litúrgica (`1ª Classe`, `2ª Classe`, `3ª Classe`, `4ª Classe`, etc.).
    3. **Calendário Mariano**: Abstinência de carne canônica (`Abstinência de carne` / `Sem abstinência de carne`), devoções de Primeira Sexta-feira e Primeiro Sábado do mês, santos congregados marianos e indulgências de Cristo Rei / Sagrado Coração.
    4. **Liturgia**: Cor litúrgica e estilização de fundo correspondente (`Branco`, `Verde`, `Vermelho`, `Roxo`, `Preto`, `Rosa`, `Azul`), linha de `Glória • Credo`, `Prefácio` e leituras sagradas (`Epístola • Evangelho`).
  - Quando solicitado em `format=xls`: serve cabeçalhos `Content-Type: application/vnd.ms-excel; charset=utf-8` e `Content-Disposition: attachment; filename="calendario_{mes}_{ano}.xls"` para abertura nativa e sem alertas em qualquer versão do Microsoft Excel ou LibreOffice.
  - Quando solicitado em `format=html`: serve `Content-Type: text/html; charset=utf-8` para renderização imediata na web e portais.
- **Documentação Interativa Scalar & OpenAPI (`api/openapi.json`)**:
  - Especificação OpenAPI 3.1 atualizada com o novo endpoint `/api/v1/calendar/export`, seus parâmetros e schemas de resposta, automaticamente refletido na UI interativa do Scalar em `/docs`.

## [2.3.0] - 2026-09-16

### 🐟 Feature: Abstinência de Carne (API / Sextas-feiras e Quarta-feira de Cinzas)
- **Cálculo Canônico de Abstinência (`engine/models.go`)**:
  - Implementada a função `CalculateAbstinence(date, day, finalName)` em conformidade com o Código de Direito Canônico (CIC 1917, c. 1252 § 1) e as Rubricas Romanas.
  - Toda sexta-feira do ano é dia de abstinência de carne (`has_abstinence: true`), **exceto** quando coincide com festa de I Classe (1962) ou Duplo de I Classe (1954/Pré-55), na qual a abstinência cessa.
  - Sexta-feira da Paixão (Sexta-feira Santa) é explicitamente preservada como dia de jejum e abstinência.
  - Quarta-feira de Cinzas é computada como dia de jejum e abstinência (`has_abstinence: true`).
  - Dias após a Quarta-feira de Cinzas (quinta-feira e sábado) são explicitamente desmarcados como dias de abstinência.
- **Exposição na API (`api/dto.go`, `api/handlers.go`, `api/openapi.json`)**:
  - Adicionado o campo `has_abstinence: bool` em `LiturgicalResponse` e em `LiturgicalDayJSON`.
  - Atualizada a especificação OpenAPI (`openapi.json`) documentando o novo campo booleano.

### 🔄 Sincronização e Correções de Calendário (LiturgyCalendarApp)
- **Correção das Têmporas de Setembro no Calendário de 1962 (`engine/temporal.go`)**:
  - Ajustado o cálculo das Têmporas de Setembro de 1962 para a primeira quarta-feira, sexta-feira e sábado após o **3º domingo de setembro**, conforme as rubricas do Código de 1960 de João XXIII (enquanto no Pré-55/1954 permanece após a Exaltação da Santa Cruz, 14 de setembro).
- **Comemoração de Têmporas com Festas no Pré-55 (`engine/profile_1954.go`)**:
  - Permitida a comemoração de dias de têmporas quando concorrem com festas (`_feriaCanBeCommemoratedWithFeast`), preservando as rubricas de 1954.
  - Ajustado o prefácio sazonal em fés apenas quando o prefácio atual for o comum (`isDefaultPreface`).
- **Sincronização de Dados Litúrgicos (`data/`)**:
  - Sincronizados `sanctoral.xml`, `temporal.xml`, `1954/sanctoral.xml`, `1954/temporal.xml`, `1954/PROVENANCE.json` e `1954/schema.json` a partir do repositório upstream `LiturgyCalendarApp`.

---

## [2.2.0] - 2026-09-14

### 📊 Observabilidade & Padrão Google SRE
- **Métricas Nativas do Prometheus (`api/metrics.go`)**:
  - Implementado o endpoint `GET /metrics` em conformidade com o padrão OpenMetrics/Prometheus em texto puro.
  - Telemetria de tráfego em tempo real categorizada por classe de status HTTP (`2xx`, `4xx`, `5xx`, `all`) e latência acumulada usando contadores atômicos (`sync/atomic`), livre de data races.
  - Exposição de métricas de runtime do Go (`go_goroutines`, `go_memstats_alloc_bytes`, `go_memstats_sys_bytes`, `go_memstats_num_gc`, `tesouro_uptime_seconds`).
  - Zero dependências externas (implementado 100% com pacotes padrão do Go).
- **Integração no LoggingMiddleware (`api/middleware.go`)**:
  - Registro de telemetria atômica em cada requisição antes do despacho de logs estruturados em JSON via `log/slog`.
- **Suíte de Benchmarks Automatizados (`engine/engine_test.go`)**:
  - `BenchmarkEasterCalculation`: Algoritmo computacional de Páscoa rodando a **12.57 ns/op** com 0 alocações.
  - `Benchmark1962Resolution`: Resolução completa do calendário litúrgico de 1962 a **1.85 µs/op** (~540k ops/seg).
  - `Benchmark1954Resolution`: Resolução pré-55 com oitavas e concorrências a **471 µs/op**.
- **Testes de Concorrência e Validação**:
  - Adicionado teste de integração `TestMetricsEndpoint` em `api/api_test.go`.
  - Verificação de concorrência estrita via `go test -race ./...` com 100% de aprovação.

---

## [2.1.0] - 2026-09-07

### 🔄 Sincronização e Correções de Calendário (LiturgyCalendarApp)
- **Próprio do Brasil de 1954 (`data/1954/brazilian_sanctoral.xml`)**:
  - Adicionada a base canônica com festas próprias brasileiras e expansão de oitavas para o rito pré-55.
  - Implementado o parser `Brazilian1954ProperCalendar` em `engine/brazilian_1954_proper.go` com suporte a oitavas comuns (ex.: São José, Sagrado Coração) e festas relativas.
  - Refinadas as precedências em `engine/profile_1954.go` e `engine/pre55_rank.go` para resolução harmoniosa entre o calendário universal e o próprio nacional.
  - Adicionado cálculo de prefácio de oitava e Credo para dias dentro de oitava.
- **Correção da Cor de Festas de Santos (`engine/models.go`)**:
  - Ajustada a verificação `isMarianFeast` para ignorar prefixos de santidade (`Santa`, `St.`, `San`, etc.), prevenindo que santas como Santa Rosa de Lima sejam incorretamente classificadas como marianas (mantendo a cor canônica `WHITE` em vez de `BLUE`).
- **Têmporas de Setembro e Advento (`engine/temporal.go`)**:
  - Corrigido o cálculo das Têmporas de Setembro para a regra tradicional pós-Exaltação da Santa Cruz (primeira quarta, sexta e sábado após 14 de setembro).
  - Adicionadas as leituras próprias de Epístola e Evangelho para os dias de Têmporas de Setembro e Advento.
- **Internacionalização (`data/values*/strings.xml`)**:
  - Atualizados os dicionários com ordinais e nomes de festas em Português (`values`, `pt-rBR`), Inglês (`en`), Espanhol (`es`), Francês (`fr`), Alemão (`de`) e Latim (`la`).

### 🚀 Automação de CI/CD e Infraestrutura Leve
- **Pipeline no GitHub Actions (`.github/workflows/deploy.yml`)**:
  - Testes unitários com detector de concorrência (`go test -v -race ./...`) em cada push e pull request.
  - Build automatizado de imagem Docker multi-arquitetura (`linux/arm64` nativo para Raspberry Pi e `linux/amd64`) via Docker Buildx com compilação cruzada nativa em Go (sem overhead de QEMU no compilador).
  - Publicação automática da imagem para o GitHub Container Registry (`ghcr.io/mathvdias/tesouro-backend-go:latest`).
  - Cache de camadas com GitHub Actions Cache (`type=gha`) para builds em poucos segundos.
- **Deploy Zero-Touch com Watchtower no Raspberry Pi**:
  - Adicionado serviço `watchtower` ao `docker-compose.yml` configurado com `--interval 300`, `--cleanup` e `--label-enable`.
  - Consumo ultraleve de recursos (~15MB RAM, ~0% CPU), sem necessidade de abrir portas SSH públicas no Raspberry Pi nem de instalar runners pesados.
  - Limpeza automática de imagens antigas após atualização (`docker image prune`).
- **Script de Deploy Instantâneo (`scripts/deploy.sh`)**:
  - Permite disparar a atualização imediata no Raspberry Pi via SSH com 1 comando, sem precisar aguardar o intervalo do Watchtower.

---

## [2.0.0] - 2026-08-26

### 🚀 Novas Funcionalidades (Dual Calendar Engine)
- **Suporte ao Calendário Litúrgico de 1954 (Divino Afflatu / Pré-55)**:
  - Adicionadas as bases de dados completas do calendário de 1954 em `data/1954/` (`temporal.xml`, `sanctoral.xml`, `schema.json`, `PROVENANCE.json`).
  - Implementado o enum `Pre55Rank` com os 6 graus de festas (*Duplex I Classis, Duplex II Classis, Duplex Maius, Duplex, Semiduplex, Simplex*) e 3 graus de férias (*Privilegiata, Major, Minor*).
  - Implementado o enum `Pre55OctaveType` com 5 tipos de oitavas (*Privilegiada 1ª, 2ª e 3ª Ordem, Comum e Simples*).
  - Implementado o motor de cálculo `DivinoAfflatu1954Profile` em Go (`engine/profile_1954.go`), incluindo:
    - Regras de concorrência com despacho de Primeiras Vésperas.
    - Transferência de festas ocorridas na Semana Santa ou Oitava da Páscoa.
    - Alocação e distribuição dinâmica dos Domingos pós-Pentecostes com reinserção dos domingos omitidos da Epifania.
    - Comemorações múltiplas em conformidade com as rubricas pré-1955.
    - Liturgia da Missa com cálculo de Glória, Credo, Prefácio, Epístola e Evangelho.
    - Sobreposição harmonizada do Próprio do Brasil (`BrazilianSanctorale`).

- **Aprimoramentos no Calendário de 1962 (Tridentino)**:
  - Atualização dos arquivos `data/temporal.xml`, `data/sanctoral.xml` e `data/brazilian_sanctoral.xml` com leituras próprias (Epístola e Evangelho).
  - Transferência automática da festa da Anunciação (25 de março) para a segunda-feira após o Domingo *in Albis* quando incidente na Semana Santa ou Oitava Pascal (Código de Rubricas de 1960, §96a).
  - Regra de propagação das leituras do domingo anterior para as férias do tempo após Pentecostes e Advento.
  - Regras estritas de supressão de comemorações em dias e tempos penitenciais (§108).
  - Missa votiva de Nossa Senhora aos Sábados (§78).
  - Cor litúrgica especial para festas marianas (`BLUE`) e de São José (`WHITE`).

- **Novos Idiomas e Traduções**:
  - Adicionado suporte completo a traduções em Latim em `data/values-la/strings.xml`.
  - Atualizados os dicionários `data/values/` (Português), `data/values-pt-rBR/`, `data/values-en/`, `data/values-es/`, `data/values-fr/` e `data/values-de/`.

- **Parâmetro de Versão do Calendário nas Rotas HTTP**:
  - `GET /api/v1/liturgical-day`: Aceita `calendar=1962` (padrão) ou `calendar=1954`.
  - `POST /api/v1/liturgical-day`: Aceita campo `"calendar": "1954"` no corpo JSON.
  - `GET /api/v1/liturgical-month`: Suporte ao parâmetro `calendar` para gerar o mês inteiro no rito escolhido.
  - 100% de compatibilidade retroativa com clientes legados que não enviam o parâmetro `calendar`.

### 🗑️ Remoções (Limpeza e Otimização)
- **Remoção Completa da Integração com Google Gemini AI e Santos Marianos**:
  - Removido o arquivo `engine/marian.go` e a struct `MarianSaint`.
  - Removidas as chamadas de rede para a API do Gemini (`generativelanguage.googleapis.com`) e os esquemas estruturados JSON.
  - Removidos os endpoints legados `/api/v1/saint-of-the-day` e `/api/v1/marian-saints`.
  - Removidas as variáveis de ambiente `GEMINI_API_KEY` e `GEMINI_MODEL` do `docker-compose.yml` e `.env`.
  - Removidos arquivos legados `temporal_cycle.xml` e `universal_sanctoral.xml`.

### 🛡️ Infraestrutura, Deploy e Testes
- Adicionada suíte abrangente de testes unitários:
  - `engine/engine_test.go`: Testes de cálculo pascal, transferências da Anunciação, festas brasileiras e precedências de 1954.
  - `main_test.go`: Testes de integração de todas as rotas HTTP e serializações JSON.
- Criada a tag de segurança e rollback `v1.0.0-1962-baseline`.
- Atualizado o guia de deploy e rollback com comandos testados em `DEPLOY.md`.
- Imagem Docker no nó bare-metal protegida com backup estável `tesouro-backend-go-liturgical-backend:backup-stable`.
