# Guia de Deploy no Raspberry Pi 🚀

Este guia documenta o fluxo de build e deploy automatizado para compilar o backend em Go (`tesouro-backend-go`) localmente no Mac e enviá-lo para execução no ambiente de produção do Raspberry Pi.

---

## Fluxo de Deploy Passo a Passo

### 1. Pré-requisitos
* O daemon Docker/Colima deve estar rodando no Mac.
* Acesso SSH configurado sem senha (via chave pública) para o usuário `matheus` no Raspberry Pi (`100.92.173.88`).

### 2. Passo 1: Construir a Imagem Docker Localmente (no Mac)
Como o Mac M-series e o Raspberry Pi compartilham a arquitetura `arm64`, podemos compilar a imagem localmente e ela rodará nativamente no Pi.

No terminal do seu Mac, na pasta raiz `tesouro-backend-go`:
```bash
# Reconstrói a imagem localmente usando Docker Compose
docker-compose up --build -d
```

### 3. Passo 2: Exportar a Imagem para um Arquivo Tar
Para transferir a imagem sem depender de um registro público (Docker Hub), salvamos a imagem como um arquivo `.tar`:
```bash
docker save -o tesouro-backend-go-liturgical-backend.tar tesouro-backend-go-liturgical-backend:latest
```

### 4. Passo 3: Enviar o Código e a Imagem para o Raspberry Pi
Usamos o `rsync` para sincronizar os arquivos de configuração, base de dados XML e a imagem exportada para a pasta `/home/matheus/tesouro-backend-go` no Pi:
```bash
rsync -avz --exclude .git --exclude .venv --exclude .DS_Store --exclude .env ./ matheus@100.92.173.88:/home/matheus/tesouro-backend-go/
```

### 5. Passo 4: Carregar a Imagem e Reiniciar a Stack no Pi
Acesse o Pi via SSH para carregar a nova imagem, derrubar os contêineres antigos (removendo órfãos) e subir a nova versão:
```bash
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker load -i tesouro-backend-go-liturgical-backend.tar && sudo docker compose down --remove-orphans && sudo docker compose up -d"
```

### 6. Passo 5: Limpeza de Arquivos Temporários
Para evitar desperdício de espaço em disco no Raspberry Pi, exclua os arquivos de código fonte e o arquivo `.tar` do sistema hospedeiro do Pi (já que tudo roda de dentro do contêiner Docker):
```bash
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && rm -rf engine/ data/ main.go Dockerfile go.mod tesouro-backend-go-liturgical-backend.tar"
```

---

## Resumo dos Comandos em Linha Única
Para fazer todo o deploy com apenas um comando no Mac:
```bash
docker-compose up --build -d && \
docker save -o tesouro-backend-go-liturgical-backend.tar tesouro-backend-go-liturgical-backend:latest && \
rsync -avz --exclude .git --exclude .venv --exclude .DS_Store --exclude .env ./ matheus@100.92.173.88:/home/matheus/tesouro-backend-go/ && \
ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker load -i tesouro-backend-go-liturgical-backend.tar && sudo docker compose down --remove-orphans && sudo docker compose up -d && rm -rf engine/ data/ main.go Dockerfile go.mod tesouro-backend-go-liturgical-backend.tar"
```

---

## Solução de Problemas

### Erro 502 Bad Gateway no Cloudflare
Se o domínio `https://api.salvemaria.xyz` retornar 502:
1. Verifique se o contêiner `cloudflared` está ativo no Pi:
   ```bash
   ssh matheus@100.92.173.88 "sudo docker ps"
   ```
2. Verifique se a variável `CLOUDFLARE_TUNNEL_TOKEN` está configurada corretamente no arquivo `/home/matheus/tesouro-backend-go/.env` no Pi.
3. Inspecione os logs do contêiner do túnel:
   ```bash
   ssh matheus@100.92.173.88 "cd /home/matheus/tesouro-backend-go && sudo docker compose logs cloudflared"
   ```
