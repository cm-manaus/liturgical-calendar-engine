# Guia de Requisições da API em Produção (SalveMaria) 🌐

Este documento contém exemplos práticos de requisições (`curl`) para os endpoints da API do Calendário Litúrgico (1962 e 1954 Pré-55) hospedados em produção.

**URL Base de Produção:** `https://api.salvemaria.xyz`

---

## 1. Status da API (Root)
Verifica a conectividade, calendários suportados e lista os endpoints ativos.

* **Método:** `GET`
* **Path:** `/`

### Exemplo de Requisição (curl):
```bash
curl -s -X GET https://api.salvemaria.xyz/
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

## 2. Calendário Litúrgico 1962 (Padrão)
Retorna o dia litúrgico segundo as rubricas de 1962 (Missal de João XXIII).

* **Métodos:** `GET` e `POST`
* **Paths:** `/api/v1/liturgical-day` e `/liturgical-day`

### Parâmetros (GET - Query Params):
* `date` *(opcional)*: Data no formato `YYYY-MM-DD`. Padrão: hoje.
* `lang` *(opcional)*: Idioma da tradução. Valores aceitos: `pt-br`, `pt`, `en`, `es`, `fr`, `de`, `la`. Padrão: `en` ou `Accept-Language`.
* `calendar` *(opcional)*: `1962` (padrão) ou `1954`.
* `include_brazilian` *(opcional)*: Exclui festas brasileiras se for `false`. Padrão: `true`.

### Exemplo de Requisição GET 1962 (curl):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-10-12&lang=pt-br"
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

## 3. Calendário Litúrgico 1954 (Divino Afflatu / Pré-55)
Retorna o dia litúrgico segundo as rubricas anteriores à reforma de 1955, incluindo graus de festa (*Duplex I/II Classis, Duplex Maius, Duplex, Semiduplex, Simplex*), tipos de oitava, comemorações múltiplas e liturgia da Missa (Gloria, Credo, Prefácio, Epístola e Evangelho).

* **Métodos:** `GET` e `POST`
* **Paths:** `/api/v1/liturgical-day?calendar=1954`

### Exemplo de Requisição GET 1954 (curl):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-01-06&calendar=1954&lang=pt-br"
```

### Exemplo de Requisição POST 1954 (curl):
```bash
curl -s -X POST https://api.salvemaria.xyz/api/v1/liturgical-day \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-01-06",
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

## 4. Calendário Mensal (Mês Completo)
Retorna a lista de todos os dias de um determinado mês em uma única chamada.

* **Método:** `GET`
* **Path:** `/api/v1/liturgical-month`

### Parâmetros (GET - Query Params):
* `year` *(opcional)*: Ano (ex: `2026`). Padrão: ano atual.
* `month` *(opcional)*: Mês de 1 a 12. Padrão: mês atual.
* `calendar` *(opcional)*: `1962` (padrão) ou `1954`.
* `lang` *(opcional)*: Idioma da tradução (`pt-br`, `pt`, `en`, `es`, `fr`, `de`, `la`).

### Exemplo de Requisição (curl):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-month?year=2026&month=8&calendar=1954&lang=pt-br"
```
