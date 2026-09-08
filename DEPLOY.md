# Guia de CI/CD, Deploy Automatizado e Monitoramento no Raspberry Pi 🚀

Este guia documenta o pipeline de integração contínua (CI/CD), publicação no GitHub Container Registry (GHCR), deploy automatizado e rollback para o backend em Go (`tesouro-backend-go`) rodando no Raspberry Pi (`100.92.173.88`) com Docker Compose e Cloudflare Tunnels (`https://api.salvemaria.xyz`).

---

## 🏗️ 1. Arquitetura do CI/CD

```
  [ Dev Mac ]
      │  git push origin main
      ▼
┌────────────────────────────────────────────────────────┐
│ GitHub Actions (.github/workflows/deploy.yml)          │
│ 1. go test -v -race ./... (Validação completa)         │
│ 2. Docker Buildx multi-arch (linux/arm64, linux/amd64) │
│    - Compilação cruzada nativa em Go (sem QEMU lento)   │
│ 3. Push para GitHub Container Registry (GHCR)          │
└──────────────────────────┬─────────────────────────────┘
                           │ ghcr.io/cm-manaus/tesouro-backend:latest
                           ▼
┌────────────────────────────────────────────────────────┐
│ Raspberry Pi (100.92.173.88)                           │
│ - Opção A (Automático): Watchtower atualiza a cada 5m  │
│ - Opção B (Instantâneo): ./scripts/deploy.sh           │
│ - Zero portas públicas expostas (Tailscale + Tunnels)  │
│ - Consumo de memória: ~7.5MB (Go) + ~12MB (Watchtower) │
└────────────────────────────────────────────────────────┘
```

---

## 🧪 2. Validação Local Antes do Push

Antes de subir commits para a branch `main`, execute a suíte de testes unitários localmente:

```bash
go test -v -race ./...
```

---

## 🚀 3. Fluxo de Deploy 100% Automatizado (Zero-Touch)

Com o pipeline configurado, você **não precisa compilar localmente, nem salvar .tar, nem fazer rsync**:

1. Faça o commit e dê push para a branch `main`:
   ```bash
   git add .
   git commit -m "feat: suas alteracoes"
   git push origin main
   ```
2. O **GitHub Actions** executa os testes, compila o container para `linux/arm64` nativamente e publica em `ghcr.io/cm-manaus/tesouro-backend:latest`.
3. O **Watchtower** no Raspberry Pi detecta a nova imagem automaticamente em até 5 minutos, atualiza o container e limpa a imagem antiga.

---

## ⚡ 4. Deploy Instantâneo (Sem Esperar o Watchtower)

Se você acabou de dar push e quer que a produção atualize imediatamente:

```bash
./scripts/deploy.sh
```

O script:
- Conecta via SSH ao Raspberry Pi.
- Executa `docker compose pull liturgical-backend` baixando a imagem nova do GHCR.
- Reinicia o serviço `liturgical-backend` em ~2 segundos.
- Remove imagens órfãs (`docker image prune -f`) para economizar o cartão SD/SSD.
- Executa um teste de fumaça na URL pública `https://api.salvemaria.xyz/`.

---

## 📊 5. Visualização de Containers e Recursos no Terminal (Substituto do Portainer)

Como o Raspberry Pi tem recursos limitados (CPU e RAM), o Portainer consumiria 150MB+ de memória RAM permanentemente. 

Em vez disso, use as ferramentas nativas do terminal que consom **0 MB** de RAM adicional:

### Ver uso de CPU e Memória em Tempo Real:
```bash
# No seu Mac, execute via SSH:
ssh matheus@100.92.173.88 "docker stats"

# Ou apenas uma foto estática:
ssh matheus@100.92.173.88 "docker stats --no-stream"
```

### Ver status e portas de todos os containers:
```bash
ssh matheus@100.92.173.88 "docker ps"
```

### Ver logs em tempo real do backend:
```bash
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && docker compose logs -f liturgical-backend"
```

---

## 🔑 6. Configuração de Acesso ao GHCR no Raspberry Pi (Executar uma única vez)

Como o repositório é privado, para que o Docker no Raspberry Pi consiga baixar a imagem do GHCR, existem duas opções:

### Opção A: Tornar o Pacote do Container Público (Mais simples e recomendado)
O código compilado não contém chaves de API nem segredos. Você pode tornar apenas o pacote Docker público:
1. Acesse: `https://github.com/orgs/cm-manaus/packages/container/package/tesouro-backend`
2. Clique em **Package settings** (barra lateral direita).
3. Role até **Danger Zone** -> **Change visibility** -> Selecione **Public**.
*Pronto! Qualquer pull funcionará sem requerer senhas no Raspberry Pi.*

### Opção B: Autenticar o Docker no Raspberry Pi com Personal Access Token (PAT)
Se preferir manter o pacote privado:
1. No GitHub: gere um Personal Access Token (Classic) com permissão `read:packages` em `https://github.com/settings/tokens`.
2. Conecte no Raspberry Pi e faça login:
   ```bash
   echo "SEU_GITHUB_PAT" | docker login ghcr.io -u Mathvdias --password-stdin
   ```

---

## 🛡️ 7. Procedimento de Rollback de Emergência

Caso uma atualização quebre em produção, reverta em segundos:

### Opção A: Reverter para a Tag Anterior do Docker (Instantâneo no Pi)
No Raspberry Pi ou via SSH, edite o `docker-compose.yml` para apontar para a tag de SHA anterior (ou tag estável) e reinicie:
```bash
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && docker compose down liturgical-backend && docker compose up -d liturgical-backend"
```

### Opção B: Reverter via Git no Mac
1. Reverta o commit no git:
   ```bash
   git revert HEAD
   git push origin main
   ```
2. O CI/CD irá compilar e aplicar a versão anterior automaticamente.

---

## 🔍 8. Verificação de Saúde Pós-Deploy (Smoke Test)

Após qualquer deploy, teste os endpoints diretamente na produção:

```bash
# 1. Status da API
curl -s https://api.salvemaria.xyz/

# 2. Calendário 1962 (Tridentino)
curl -s "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-08-26&lang=pt-br"

# 3. Calendário 1954 (Divino Afflatu / Pré-55 com Próprio do Brasil)
curl -s "https://api.salvemaria.xyz/api/v1/liturgical-day?date=2026-08-26&calendar=1954&lang=pt-br"
```
