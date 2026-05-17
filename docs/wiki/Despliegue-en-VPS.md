# Despliegue en VPS Ubuntu

Guía paso a paso para servir SkiHub en un VPS limpio (Ubuntu 22.04+).
Resultado final: `https://midominio.com` → Nginx (HTTPS) → binario Go (systemd) → Postgres.

## Requisitos previos

- VPS con Ubuntu 22.04 LTS o superior y acceso SSH.
- Dominio apuntando con un registro `A` a la IP del VPS.
- Puertos abiertos: 22 (SSH), 80, 443.

## 1. Usuario de despliegue

```bash
# Como root
adduser deploy
usermod -aG sudo deploy
mkdir -p /home/deploy/.ssh
cp ~/.ssh/authorized_keys /home/deploy/.ssh/
chown -R deploy:deploy /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
```

A partir de aquí, todo como `deploy`.

## 2. Dependencias del sistema

```bash
sudo apt update
sudo apt install -y git nginx postgresql ufw certbot python3-certbot-nginx

# Go (versión coincidente con go.mod, ajusta el número)
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh
go version
```

## 3. Postgres

```bash
sudo -u postgres psql <<SQL
CREATE USER skihub WITH PASSWORD 'CAMBIA_ESTO_POR_ALGO_LARGO';
CREATE DATABASE skihub OWNER skihub;
SQL
```

## 4. Firewall

```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
sudo ufw status
```

## 5. Clonar el repo y compilar

```bash
sudo mkdir -p /var/www && sudo chown deploy:deploy /var/www
cd /var/www
git clone https://github.com/celadaa/web_sky.git skihub
cd skihub
cp .env.example .env
nano .env   # Rellena DB_*, COOKIE_SECRET, etc.

mkdir -p bin
go build -o bin/skihub ./cmd/servidor
./bin/skihub --version   # smoke test si lo soportas, si no Ctrl+C tras unos segundos
```

## 6. Servicio systemd

Crea `/etc/systemd/system/skihub.service`:

```ini
[Unit]
Description=SkiHub web server
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=deploy
WorkingDirectory=/var/www/skihub
EnvironmentFile=/var/www/skihub/.env
ExecStart=/var/www/skihub/bin/skihub
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

# Sandbox: arnés básico contra escapes desde el binario
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=read-only
ReadWritePaths=/var/www/skihub

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now skihub
sudo systemctl status skihub
journalctl -u skihub -f          # logs en directo
```

## 7. Nginx + HTTPS

Copia `deploy/nginx/snowbreak` a `/etc/nginx/sites-available/skihub` y ajústalo a tu dominio.

```bash
sudo ln -s /etc/nginx/sites-available/skihub /etc/nginx/sites-enabled/skihub
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx

# HTTPS con Let's Encrypt
sudo certbot --nginx -d midominio.com -d www.midominio.com
sudo systemctl reload nginx
```

`certbot` instala una tarea de renovación automática (cron/systemd timer).

## 8. Health endpoint

Asegúrate de exponer `/healthz` que devuelva `200 OK` sin tocar la BD. El workflow
de despliegue lo usa para validar tras cada release.

## 9. Despliegue automático (GitHub Actions)

Una vez todo funciona a mano, configura el workflow:

1. En el VPS: `ssh-keygen -t ed25519 -f ~/.ssh/github_deploy -N ""`.
2. `cat ~/.ssh/github_deploy.pub >> ~/.ssh/authorized_keys`.
3. En GitHub → Settings → Secrets and variables → Actions, crea:
   - `VPS_HOST`, `VPS_USER` (= `deploy`), `VPS_PORT` (= `22`)
   - `VPS_SSH_KEY` = contenido completo de `~/.ssh/github_deploy` (privada)
   - `VPS_APP_PATH` = `/var/www/skihub`
   - `VPS_SERVICE` = `skihub.service`
   - `HEALTH_URL` = `https://midominio.com/healthz`
4. Da permiso a `deploy` para reiniciar el servicio sin contraseña:
   ```bash
   echo 'deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart skihub.service, /bin/systemctl is-active skihub.service, /bin/systemctl status skihub.service' | sudo tee /etc/sudoers.d/deploy-skihub
   sudo chmod 440 /etc/sudoers.d/deploy-skihub
   ```
5. Lanza un push a `main` y observa la pestaña Actions.

## Rollback manual

```bash
ssh deploy@VPS
cd /var/www/skihub
cat .last_deployed_sha          # SHA anterior, lo guarda el workflow
git checkout <SHA>
go build -o bin/skihub ./cmd/servidor
sudo systemctl restart skihub
```
