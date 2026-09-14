# ⚜️ tesouro-backend-go

[![CI/CD Pipeline](https://github.com/cm-manaus/tesouro-backend/actions/workflows/deploy.yml/badge.svg)](https://github.com/cm-manaus/tesouro-backend/actions/workflows/deploy.yml)
![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8.svg?logo=go)
![Observability](https://img.shields.io/badge/Prometheus-ready-E6522C.svg?logo=prometheus)
![Resolution](https://img.shields.io/badge/Resolution-1.85µs-brightgreen)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

Este repositório contém o **Motor Litúrgico Tradicional (1962 e 1954 Pré-55)** em Go que alimenta a plataforma **Salve Maria**. O projeto foi estruturado para obter máxima performance, baixíssimo consumo de recursos e segurança em contêineres Docker no Raspberry Pi e clusters bare-metal.

---

## 🗺️ Arquitetura do Sistema e Design de Infraestrutura

O diagrama abaixo ilustra o ciclo de vida completo: esteira de CI/CD automatizada com compilação multi-arquitetura, distribuição contínua para o nó bare-metal, exposição Zero Trust via Cloudflare Tunnels e coleta de métricas em tempo real com Prometheus:

```mermaid
graph TD
    %% Estilo dos Nós
    classDef dev fill:#2b303a,stroke:#4a5568,stroke-width:2px,color:#fff;
    classDef server fill:#1a202c,stroke:#3182ce,stroke-width:2px,color:#fff;
    classDef container fill:#2d3748,stroke:#38b2ac,stroke-width:2px,color:#fff;
    classDef cloud fill:#2d3748,stroke:#ed8936,stroke-width:2px,color:#fff;
    classDef client fill:#2d3748,stroke:#4a5568,stroke-width:2px,color:#fff;
    classDef obs fill:#1f2937,stroke:#e6522c,stroke-width:2px,color:#fff;

    subgraph dev_pipeline ["🚀 Esteira de CI/CD (GitHub Actions)"]
        A["Git Push (main)"]:::dev
        B["Testes Automatizados<br/>• go test -v -race<br/>• govulncheck"]:::dev
        C["Docker Buildx Multi-Arch<br/>(linux/arm64, linux/amd64)"]:::dev
        D["GitHub Container Registry<br/>(ghcr.io/cm-manaus/tesouro-backend:latest)"]:::dev

        A --> B --> C --> D
    end

    subgraph cloudflare_edge ["☁️ Borda do Cloudflare"]
        I["DNS & Proxy (api.salvemaria.xyz)"]:::cloud
        J["Filtros WAF / DDoS Mitigation"]:::cloud
        K["Cloudflare Edge Network"]:::cloud
        
        I --> J --> K
    end

    subgraph home_server ["🍓 Bare-Metal Node (Raspberry Pi / OptiPlex)"]
        subgraph docker_compose ["Docker Compose Stack"]
            F["liturgical-backend (Go App)<br/>• Port: 8080<br/>• Latência: ~1.85µs/op"]:::container
            G["cloudflared (Cloudflare Tunnel Client)"]:::container
            W["Watchtower (Automated CD)"]:::container
            
            G <-->|Conexão Segura Interna| F
            W -.->|Pull & Deploy Contínuo| D
            W -.->|Atualização Automática| F
        end

        subgraph observability ["📊 Observabilidade SRE"]
            P["Prometheus Engine"]:::obs
            H["Grafana Dashboard"]:::obs
            
            P -->|Scrape GET /metrics a cada 15s| F
            H -->|Query Métricas & Alertas| P
        end
    end

    %% Client Request Flow
    L["Clientes & Aplicativos Mobile"]:::client -->|GET /api/v1/liturgical-day| I
    K <-->|"Túnel Reverso Seguro (Zero Trust / Sem Port Forwarding)"| G
```

---

## ⚡ Suporte a Calendários Litúrgicos

1. **1962 (Missale Romanum de João XXIII):**
   - Sistema de 4 Classes (`I`, `II`, `III`, `IV`).
   - Transferência da Anunciação (§96a) quando incidente na Semana Santa ou Oitava Pascal.
   - Regras de supressão de comemorações penitenciais (§108).
   - Missa de Nossa Senhora aos Sábados (§78).
   - Metadados litúrgicos: Glória, Credo, Prefácio, Epístola e Evangelho.

2. **1954 (Divino Afflatu / Pré-55):**
   - Sistema de 6 Graus de Festas (*Duplex I Classis, Duplex II Classis, Duplex Maius, Duplex, Semiduplex, Simplex*) e 3 Graus de Férias (*Privilegiata, Major, Minor*).
   - Oitavas pré-55 (Privilegiada de 1ª, 2ª, 3ª Ordem, Comum e Simples).
   - Regras completas de concorrência (Primeiras Vésperas), ocorrência e comemorações múltiplas.
   - Distribuição dinâmica de Domingos após Pentecostes com realocação dos domingos omitidos da Epifania.
   - Regras de transferências de festas da Semana Santa e Oitava Pascal.
   - Próprio do Brasil sobreposto de forma harmonizada.

---

## 🔒 Hardening e Segurança de Contêineres

Para rodar em ambiente de produção com segurança, o contêiner do backend no `docker-compose.yml` aplica as seguintes regras de segurança da kernel Linux:

1. **`read_only: true`:** O contêiner não pode alterar nenhum arquivo dentro de sua imagem compilada.
2. **`tmpfs: - /tmp`:** Diretório temporário montado diretamente na RAM para operações temporárias necessárias do Go.
3. **`cap_drop: - ALL`:** Remove todas as capacidades especiais da kernel Linux do processo.
4. **`security_opt: - no-new-privileges:true`:** Previne que processos filhos ganhem mais privilégios que o processo pai.

---

## 📦 Estrutura de Arquivos do Projeto

```bash
├── main.go               # Ponto de entrada e rotas HTTP (ServeMux)
├── main_test.go          # Testes de integração das rotas HTTP
├── Dockerfile            # Dockerfile multi-stage minimalista (Alpine ~8.2MB)
├── docker-compose.yml    # Definição dos serviços do backend e do túnel cloudflared
├── DEPLOY.md             # Guia de deploy e procedimento de rollback de emergência
├── api_requests_prod.md  # Exemplos curl de todos os endpoints
├── data/                 # Bases de dados XML e arquivos de tradução
│   ├── 1954/             # Temporal e Sanctoral do calendário de 1954 (Divino Afflatu)
│   ├── temporal.xml      # Ciclo temporal de 1962 com leituras
│   ├── sanctoral.xml     # Santoral universal de 1962 com leituras
│   ├── brazilian_sanctoral.xml # Santoral do Brasil com leituras e festas próprias
│   └── values-*/         # Dicionários de tradução (pt, pt-BR, en, es, fr, de, la)
└── engine/               # Motor de cálculo litúrgico
    ├── easter.go         # Algoritmo de computação da data da Páscoa
    ├── engine.go         # Orquestrador dual de calendários (1962 e 1954)
    ├── engine_test.go    # Testes unitários do motor litúrgico
    ├── localization.go   # Gerenciador de traduções Android XML
    ├── models.go         # Modelos de domínio e serialização JSON
    ├── pre55_rank.go     # Enums e regras de precedência pré-55
    ├── profile_1954.go   # Motor do calendário de 1954
    ├── sanctorale.go     # Santoral de 1962 e Próprio do Brasil
    └── temporal.go       # Temporal de 1962
```

---

## 📡 Guia de Requisições cURL da API

A API está disponível publicamente em **`https://api.salvemaria.xyz`** (ou `http://localhost:8080` em desenvolvimento local).

### 1. Status & Metadados
```bash
curl -s -X GET "https://api.salvemaria.xyz/"
```

### 2. Dia Litúrgico (1962 - Tridentino)
```bash
# Dia atual em português
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?lang=pt-br"

# Data específica (com leituras da Missa)
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-10-12&calendar=1962&lang=pt-br"

# Via POST com JSON
curl -s -X POST "https://api.salvemaria.xyz/api/v1/liturgical-day" \
  -H "Content-Type: application/json" \
  -d '{"date": "2026-12-25", "calendar": "1962", "lang": "pt-br"}'
```

### 3. Dia Litúrgico (1954 - Divino Afflatu / Pré-55)
```bash
# Dia atual pré-55 com graus clássicos (Duplex, Semiduplex, Simplex)
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?calendar=1954&lang=pt-br"

# Festa com Oitava e Leituras
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-01-06&calendar=1954&lang=pt-br"
```

### 4. Mês Completo
```bash
# Mês atual (1962)
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?lang=pt-br"

# Mês específico em 1954 (Agosto de 2026)
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?year=2026&month=8&calendar=1954&lang=pt-br"
### 5. Métricas e Observabilidade SRE (Prometheus)
```bash
# Scrape de métricas em texto puro padrão OpenMetrics/Prometheus
curl -s -X GET "https://api.salvemaria.xyz/metrics"
```

*Para a lista completa de parâmetros e exemplos de respostas JSON, consulte o [`api_requests_prod.md`](api_requests_prod.md).*

---

## 🚀 Como fazer o Deploy e Testar

### 1. Testes e Benchmarks Locais
```bash
# Executar todos os testes com detector de race conditions
go test -v -race ./...

# Executar benchmarks de resolução e consumo de memória
go test -bench=. -benchmem ./engine
```

### 2. Deploy Automatizado (GitOps / CI/CD)
O deploy é 100% automatizado via GitHub Actions. Qualquer push na branch `main` executa a suíte de testes, gera a imagem multi-arch e distribui continuamente:
```bash
git push origin main
```

### 3. Execução Local ou Self-Hosted via Docker Compose
```bash
docker compose up -d
```

*Para instruções completas de rollback de emergência, consulte o [`DEPLOY.md`](DEPLOY.md).*

