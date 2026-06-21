# Guia de Requisições da API em Produção (SalveMaria) 🌐

Este documento contém exemplos práticos de requisições (`curl`) para os endpoints da API do Calendário Litúrgico e Santo do Dia hospedados em produção.

**URL Base de Produção:** `https://api.salvemaria.xyz`

---

## 1. Status da API (Root)
Verifica a conectividade e retorna a lista de endpoints ativos.

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
  "endpoints": {
    "liturgical_day": "/api/v1/liturgical-day",
    "marian_saints": "/api/v1/marian-saints",
    "saint_of_the_day": "/api/v1/saint-of-the-day"
  }
}
```

---

## 2. Calendário Litúrgico (Dia Litúrgico)
Retorna a classificação litúrgica de 1962 de uma data específica (Cor, Classe, Solenidades e Comemorações), incluindo o Próprio do Brasil por padrão.

* **Métodos:** `GET` e `POST`
* **Paths:** `/api/v1/liturgical-day` e `/liturgical-day`

### Parâmetros (GET - Query Params):
* `date` *(opcional)*: Data no formato `YYYY-MM-DD`. Padrão: data local de hoje.
* `lang` *(opcional)*: Idioma da tradução. Valores aceitos: `pt-br`, `pt`, `en`, `es`, `fr`, `de`. Padrão: usa o cabeçalho `Accept-Language` ou cai para `en`.
* `include_brazilian` *(opcional)*: Exclui festas brasileiras se for `false`. Padrão: `true`.

### Exemplo de Requisição GET (curl):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-06-21&lang=pt-br"
```

### Exemplo de Requisição POST (curl):
```bash
curl -s -X POST https://api.salvemaria.xyz/api/v1/liturgical-day \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-06-21",
    "lang": "pt-br",
    "include_brazilian": true
  }'
```

### Exemplo de Resposta (JSON):
```json
{
  "main_day": {
    "name": "4º Domingo depois de Pentecostes",
    "class_code": "II",
    "class_name": "II Classe",
    "color": "GREEN",
    "is_lord_feast": false
  },
  "commemorations": [
    {
      "name": "S. Luís Gonzaga, Confessor",
      "class_code": "III",
      "class_name": "III Classe",
      "color": "WHITE",
      "is_lord_feast": false
    }
  ],
  "date": "2026-06-21",
  "requested_lang": "pt-br",
  "resolved_lang": "pt-br",
  "include_brazilian": true
}
```

---

## 3. Santo do Dia (Congregados Marianos com IA)
Sorteia o Santo do Dia da base estática dos 126 congregados marianos, chamando a IA (Gemini 2.5 Flash) para formatar a biografia e sugerir uma resolução espiritual em português.

* **Métodos:** `GET` e `POST`
* **Path:** `/api/v1/saint-of-the-day`

### Parâmetros (GET - Query Params):
* `date` *(opcional)*: Data no formato `YYYY-MM-DD`. Padrão: hoje.
* `exclude` *(opcional, repetível)*: Nome exato de santos a serem excluídos do sorteio (útil para evitar repetição).
* `force_saint` *(opcional)*: Nome (completo ou parte) de um santo para gerar sua biografia imediatamente.

### Exemplo de Requisição GET (curl):
```bash
curl -s -X GET "https://api.salvemaria.xyz/api/v1/saint-of-the-day?date=2026-06-21"
```

### Exemplo de Requisição GET com Exclusão e Forçamento de Santo:
```bash
# Excluindo São Luís Gonzaga
curl -s -X GET "https://api.salvemaria.xyz/api/v1/saint-of-the-day?date=2026-06-21&exclude=S.%20Lu%C3%ADs%20Gonzaga"

# Forçando a biografia de São José de Anchieta
curl -s -X GET "https://api.salvemaria.xyz/api/v1/saint-of-the-day?force_saint=Anchieta"
```

### Exemplo de Requisição POST (curl):
```bash
curl -s -X POST https://api.salvemaria.xyz/api/v1/saint-of-the-day \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-06-21",
    "exclude": ["S. Luís Gonzaga", "S. Afonso Rodrigues"],
    "force_saint": ""
  }'
```

### Exemplo de Resposta (JSON):
```json
{
  "date": "2026-06-21",
  "saint": {
    "name": "S. Luís Gonzaga",
    "feast": "21/jun",
    "context": "Confessor, declarado pela Igreja 'Padroeiro da Juventude'. Congregado em Roma, Itália."
  },
  "biography": "São Luís Gonzaga, nascido Luigi Gonzaga em 9 de março de 1568...",
  "resolution": "Inspirado pela pureza, desapego e serviço de São Luís Gonzaga...",
  "is_fallback": false
}
```

---

## 4. Base de Dados Completa dos Santos Marianos
Retorna a lista JSON completa de todos os 126 Santos e Beatos Congregados Marianos cadastrados localmente no backend. Útil para o cliente Flutter baixar em cache e posicionar marcadores de festa nos dias do calendário de forma offline.

* **Método:** `GET`
* **Path:** `/api/v1/marian-saints`

### Exemplo de Requisição (curl):
```bash
curl -s -X GET https://api.salvemaria.xyz/api/v1/marian-saints
```

### Exemplo de Resposta (JSON):
```json
[
  {
    "name": "S. Afonso Maria de Liguori",
    "feast": "2/ago",
    "context": "Bispo e Doutor da Igreja. Fundador dos Missionários Redentoristas, grande incentivador das CC.MM. Congregado em Nápoles, Itália."
  },
  {
    "name": "S. Afonso Rodrigues",
    "feast": "30/out",
    "context": "Confessor e Religioso (jesuíta). Diretor de CM na ilha de Maiorca, Espanha."
  }
  // ... mais 124 santos
]
```
