# 📜 Changelog - tesouro-backend-go

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/), e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

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
- Imagem Docker no Raspberry Pi (`100.92.173.88`) protegida com backup estável `tesouro-backend-go-liturgical-backend:backup-stable`.
