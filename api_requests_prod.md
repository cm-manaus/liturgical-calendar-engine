# Guia Completo de Requisições da API em Produção (SalveMaria) 🌐

Este documento contém exemplos práticos de requisições (`curl`) para todos os endpoints da API do Calendário Litúrgico (1962 e 1954 Pré-55) hospedados em produção.

**URL Base de Produção:** `https://api.salvemaria.xyz`  
**URL Base Local (Dev):** `http://localhost:8080`

---

## 1. Status da API & Metadados (Root)
Verifica a integridade do serviço, lista as rotas ativas e os calendários disponíveis.

* **Método:** `GET`
* **Path:** `/`

### Requisição cURL:
```bash
curl -s -X GET "https://api.salvemaria.xyz/"
```

### Exemplo de Resposta (JSON):
```json
{
  "message": "Welcome to the Go Liturgical Day API",
  "docs_url": "/docs",
  "calendars": [
    {
      "id": "1962",
      "name": "1962 (Tridentine)"
    },
    {
      "id": "1954",
      "name": "1954 (Divino Afflatu / Pre-55)"
    }
  ],
  "endpoints": {
    "liturgical_day": "/api/v1/liturgical-day",
    "liturgical_month": "/api/v1/liturgical-month"
  }
}
```

---

## 2. Dia Litúrgico — Calendário de 1962 (Tridentino / João XXIII)
Calcula o dia litúrgico segundo as rubricas de 1960/1962 (sistema de 4 classes).

* **Métodos:** `GET` e `POST`
* **Paths:** `/api/v1/liturgical-day` e `/liturgical-day`

### A. Dia Atual (Padrão em Português):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?lang=pt-br"
```

### B. Data Específica com Leituras (Glória, Credo, Epístola e Evangelho):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-10-12&calendar=1962&lang=pt-br"
```

### C. Consulta via `POST` com JSON:
```bash
curl -s -X POST "https://api.salvemaria.xyz/api/v1/liturgical-day" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-12-25",
    "calendar": "1962",
    "lang": "pt-br",
    "include_brazilian": true
  }'
```

### Exemplo de Resposta (JSON):
```json
{
  "main_day": {
    "name": "Nossa Senhora da Conceição Aparecida, Padroeira do Brasil",
    "id": "our_lady_aparecida",
    "name_res_id": "our_lady_aparecida",
    "observance_key": "res:our_lady_aparecida:",
    "calendar_version": "tridentine_1962",
    "calendar_name": "1962",
    "class_code": "I",
    "class_name": "I Classe",
    "color": "BLUE",
    "is_lord_feast": false,
    "date": "10-12",
    "liturgy": {
      "gloria": "Glória",
      "credo": "Credo",
      "preface": "Prefácio de Nossa Senhora",
      "epistle": "Eclo 24,17-21",
      "gospel": "Lc 11,27-28"
    }
  },
  "commemorations": [],
  "date": "2026-10-12",
  "requested_lang": "pt-br",
  "resolved_lang": "pt-br",
  "calendar_version": "tridentine_1962",
  "calendar_name": "1962",
  "include_brazilian": true
}
```

---

## 3. Dia Litúrgico — Calendário de 1954 (Divino Afflatu / Pré-55)
Calcula o dia litúrgico segundo as rubricas clássicas de São Pio X com graus de festa (*Duplex I/II Classis, Duplex Maius, Duplex, Semiduplex, Simplex*), tipos de oitava, regras de concorrência e comemorações múltiplas.

* **Métodos:** `GET` e `POST`
* **Paths:** `/api/v1/liturgical-day?calendar=1954`

### A. Dia Atual no Pré-55:
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?calendar=1954&lang=pt-br"
```

### B. Solenidade com Oitava Pré-55 (Ex: Epifania):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-01-06&calendar=1954&lang=pt-br"
```

### C. Consulta via `POST` com JSON:
```bash
curl -s -X POST "https://api.salvemaria.xyz/api/v1/liturgical-day" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-08-15",
    "calendar": "1954",
    "lang": "pt-br",
    "include_brazilian": true
  }'
```

### Exemplo de Resposta 1954 (JSON):
```json
{
  "main_day": {
    "name": "Epifania de Nosso Senhor Jesus Cristo",
    "id": "the_epiphany_of_our_lord",
    "name_res_id": "epiphany",
    "observance_key": "res:epiphany:",
    "calendar_version": "divino_afflatu_1954",
    "calendar_name": "1954 (Divino Afflatu)",
    "pre55_grade": "D1Cl",
    "rank_code": "D1Cl",
    "rank_name": "Duplex I Classis",
    "octave_type_name": "Oitava Privilegiada de 2ª Ordem",
    "color": "WHITE",
    "is_lord_feast": true,
    "date": "01-06",
    "liturgy": {
      "gloria": "Glória",
      "credo": "Credo",
      "preface": "Prefácio da Epifania",
      "epistle": "Is 60,1-6",
      "gospel": "Mt 2,1-12"
    },
    "observance_kind": "feast",
    "season": "Epiphany",
    "privileged": false,
    "precedence": 6.5,
    "first_vespers": true,
    "occurrence": "commemorate_or_transfer",
    "concurrence": "first_vespers",
    "octave_id": "Epiphany",
    "octave_day": 0,
    "octave_status": "day_within",
    "transfer_status": "none"
  },
  "commemorations": [],
  "date": "2026-01-06",
  "requested_lang": "pt-br",
  "resolved_lang": "pt-br",
  "calendar_version": "divino_afflatu_1954",
  "calendar_name": "1954 (Divino Afflatu)",
  "include_brazilian": true
}
```

---

## 4. Mês Litúrgico Completo (`/api/v1/liturgical-month`)
Retorna a lista de todos os dias de um determinado mês em uma única requisição com alto desempenho.

* **Método:** `GET`
* **Path:** `/api/v1/liturgical-month`

### A. Mês Atual em 1962:
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?lang=pt-br"
```

### B. Mês Específico no Pré-55 (Ex: Agosto de 2026 em 1954):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?year=2026&month=8&calendar=1954&lang=pt-br"
```

### C. Mês em Latim Clássico (`lang=la`):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?year=2026&month=8&calendar=1962&lang=la"
```

---

## 5. Tabela Completa de Parâmetros

| Parâmetro | Local | Tipo | Valores Aceitos | Padrão | Descrição |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `calendar` | Query / JSON | `string` | `1962`, `1954` | `1962` | Motor de cálculo das rubricas litúrgicas |
| `date` | Query / JSON | `string` | `YYYY-MM-DD` | Dia atual | Data para a consulta do dia litúrgico |
| `year` | Query | `int` | `1900` a `2100` | Ano atual | Ano para a consulta do mês litúrgico |
| `month` | Query | `int` | `1` a `12` | Mês atual | Mês para a consulta do mês litúrgico |
| `lang` | Query / JSON | `string` | `pt-br`, `pt`, `en`, `es`, `fr`, `de`, `la` | `pt-br` ou `en` | Idioma de tradução dos nomes e festas |
| `include_brazilian`| Query / JSON | `bool` | `true`, `false` | `true` | Inclui festas e solenidades próprias do Brasil |
