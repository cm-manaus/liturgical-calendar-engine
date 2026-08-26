# Guia de Deploy e Rollback no Raspberry Pi 🚀

Este guia documenta o fluxo de compilação, teste, deploy automatizado e rollback de emergência para o backend em Go (`tesouro-backend-go`) rodando no Raspberry Pi (`100.92.173.88`) sob Docker Compose e Cloudflare Tunnels (`https://api.salvemaria.xyz`).

---

## 🧪 1. Validação Local Antes do Deploy

Antes de enviar qualquer build para a produção, execute a suíte de testes unitários localmente:

```bash
# Executa todos os testes unitários do motor litúrgico e rotas HTTP
go test -v ./...
```

---

## 🚀 2. Fluxo de Deploy Passo a Passo

### Pré-requisitos
* Daemon Docker/Colima rodando no Mac.
* Acesso SSH configurado para `matheus@100.92.173.88`.

### Passo 1: Construir a Imagem Localmente
```bash
docker compose build
```

### Passo 2: Exportar a Imagem para Arquivo Tar
```bash
docker save -o tesouro-backend-go-liturgical-backend.tar tesouro-backend-go-liturgical-backend:latest
```

### Passo 3: Enviar a Imagem e Configurações para o Pi
```bash
rsync -avz --exclude .git --exclude .venv --exclude .DS_Store --exclude .env ./ matheus@100.92.173.88:/home/matheus/tesouro-backend-go/
```

### Passo 4: Carregar a Imagem e Reiniciar a Stack no Pi
```bash
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker load -i tesouro-backend-go-liturgical-backend.tar && sudo docker compose down --remove-orphans && sudo docker compose up -d && rm -rf engine/ data/ main.go Dockerfile go.mod tesouro-backend-go-liturgical-backend.tar"
```

---

## ⚡ Comando Unificado de Deploy (Linha Única)

Execute no Mac:
```bash
go test ./... && \
docker compose build && \
docker save -o tesouro-backend-go-liturgical-backend.tar tesouro-backend-go-liturgical-backend:latest && \
rsync -avz --exclude .git --exclude .venv --exclude .DS_Store --exclude .env ./ matheus@100.92.173.88:/home/matheus/tesouro-backend-go/ && \
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker load -i tesouro-backend-go-liturgical-backend.tar && sudo docker compose down --remove-orphans && sudo docker compose up -d && rm -rf engine/ data/ main.go Dockerfile go.mod tesouro-backend-go-liturgical-backend.tar"
```

---

## 🛡️ 3. Procedimento de Rollback de Emergência

Caso a nova versão apresente qualquer comportamento inesperado em produção, siga este procedimento de rollback:

### Opção A: Rollback via Git Tag no Mac
O ponto estável anterior está gravado na tag git `v1.0.0-1962-baseline`.

1. Volte o código para a tag estável:
   ```bash
   git checkout v1.0.0-1962-baseline
   ```
2. Recompile e faça o deploy imediato:
   ```bash
   docker compose build && \
   docker save -o tesouro-backend-go-liturgical-backend.tar tesouro-backend-go-liturgical-backend:latest && \
   rsync -avz --exclude .git --exclude .venv --exclude .DS_Store --exclude .env ./ matheus@100.92.173.88:/home/matheus/tesouro-backend-go/ && \
   ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker load -i tesouro-backend-go-liturgical-backend.tar && sudo docker compose down --remove-orphans && sudo docker compose up -d && rm -rf engine/ data/ main.go Dockerfile go.mod tesouro-backend-go-liturgical-backend.tar"
   ```
3. Retorne à branch `main`:
   ```bash
   git checkout main
   ```

---

## 🔍 4. Verificação de Saúde Pós-Deploy (Smoke Test)

Após o deploy, teste os endpoints diretamente na produção:

```bash
# 1. Status da API
curl -s https://api.salvemaria.xyz/

# 2. Calendário 1962 (Padrão)
curl -s "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-08-26&lang=pt-br"

# 3. Calendário 1954 (Divino Afflatu / Pré-55)
curl -s "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-08-26&calendar=1954&lang=pt-br"
```
