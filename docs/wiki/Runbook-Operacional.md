# Runbook operacional

Qué hacer cuando algo va mal. Pensado para ti mismo dentro de 3 meses cuando no recuerdes.

## Diagnóstico rápido

```bash
ssh deploy@VPS
sudo systemctl status skihub        # ¿corre el servicio?
sudo systemctl status nginx
sudo systemctl status postgresql
journalctl -u skihub -n 100 --no-pager
sudo tail -n 100 /var/log/nginx/error.log
df -h                               # ¿queda disco?
free -m                             # ¿queda RAM?
```

## Escenarios

### 1. La web no responde (timeout / 502)

```bash
sudo systemctl status skihub
journalctl -u skihub -n 200 --no-pager | tail -50
```

- Si el proceso está caído: `sudo systemctl restart skihub`.
- Si `bind: address already in use`: hay otra instancia. `sudo ss -tlnp | grep 8080`,
  mata la anterior con `kill <PID>` y reinicia.
- Si arranca y se cae al instante: variables de entorno mal en `.env`,
  o BD inaccesible. Revisa logs.

### 2. La BD no acepta conexiones

```bash
sudo systemctl status postgresql
sudo -u postgres psql -c "SELECT now();"
sudo tail -n 100 /var/log/postgresql/postgresql-*-main.log
```

- Disco lleno: `df -h /var/lib/postgresql`. Vacía logs viejos o reduce retención.
- `too many connections`: hay leaks. Reinicia skihub
  (`sudo systemctl restart skihub`) como mitigación y abre issue con stacktrace.

### 3. Despliegue ha roto producción

Rollback rápido:

```bash
cd /var/www/skihub
PREV=$(cat .last_deployed_sha)
git checkout "$PREV"
go build -o bin/skihub ./cmd/servidor
sudo systemctl restart skihub
curl -fsS https://midominio.com/healthz
```

Luego, reverte el commit en GitHub y abre un issue post-mortem.

### 4. Certificado HTTPS expirado

```bash
sudo certbot renew
sudo systemctl reload nginx
```

Si falla la renovación, revisa que el puerto 80 esté abierto y que el dominio
sigue apuntando al VPS.

### 5. Disco lleno

```bash
df -h
du -sh /var/log/* | sort -h | tail
sudo journalctl --vacuum-size=200M
sudo apt clean
```

Si el problema es Postgres: revisa `pg_wal` y considera `WAL archiving` o un
`VACUUM FULL` planificado en ventana de baja carga.

### 6. Picos de tráfico / DDoS

- Nginx ya tiene `limit_req` configurado (revisa `deploy/nginx/snowbreak`).
- Si vuelve a pasar: activa Cloudflare/BunnyCDN por delante en modo proxy.

## Backups

> ⚠️ **Antes de tocar nada en producción**, valida que tienes un backup reciente.

```bash
# Backup manual
pg_dump -U skihub -h 127.0.0.1 -F c skihub > "skihub_$(date +%F).dump"

# Restore
pg_restore -U skihub -h 127.0.0.1 -d skihub_new --clean --if-exists skihub_2026-05-17.dump
```

Recomendación: `cron` diario que sube el dump a almacenamiento externo (S3, B2, Hetzner Storage Box).
Mira [Roadmap](Roadmap) para el TODO de automatizar esto.

## Contactos / dependencias externas

- DNS: `<proveedor de dominio>`
- VPS: `<proveedor>`
- Let's Encrypt: gratuito, sin contacto, renueva solo.
- Email transaccional (si lo añades): `<servicio>`
