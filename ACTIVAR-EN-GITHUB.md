# 🚀 Activar todo lo nuevo en GitHub

Esta guía te lleva paso a paso por las cosas que **no se pueden configurar
desde el repo** (todo lo demás ya quedó listo al hacer commit).
Tiempo total estimado: **15–20 minutos**.

---

## 1. Commitea y empuja los cambios

```bash
git add .github/ docs/ SECURITY.md CONTRIBUTING.md CHANGELOG.md \
        .golangci.yml .editorconfig Makefile README.md ACTIVAR-EN-GITHUB.md
git commit -m "chore: añadir CI/CD, seguridad, plantillas y wiki

- Workflows: ci, deploy a VPS, codeql, dependency-review, gitleaks
- Dependabot semanal para go modules, actions y docker
- Plantillas de issues, PR y CODEOWNERS
- Documentación: SECURITY, CONTRIBUTING, CHANGELOG y wiki en docs/wiki
- Calidad: .golangci.yml, .editorconfig, Makefile"
git push origin main
```

Al hacer push, los workflows que NO requieren secrets empiezan a correr solos:
**CI**, **CodeQL**, **Dependency Review**, **Gitleaks**.

---

## 2. ⚙️ Settings → General

- ✅ **Features**:
  - Marca **Issues** (necesario para las plantillas).
  - Marca **Wiki** y deja el acceso restringido a colaboradores.
  - Marca **Discussions** (si quieres separar preguntas de bugs).
  - Marca **Projects**.

- ✅ **Pull Requests**:
  - **Allow squash merging**. Desmarca los otros (historial limpio).
  - **Automatically delete head branches** ✅.

---

## 3. 🔐 Settings → Branches → Branch protection rules

Crea una regla para `main`:

- ✅ **Require a pull request before merging**
  - ✅ Require approvals: 1 (o 0 si trabajas solo, pero mejor 1 con auto-approve de tu PR).
  - ✅ Dismiss stale approvals on new commits.
  - ✅ Require review from Code Owners.
- ✅ **Require status checks to pass before merging**
  - Marca: `Build, test & lint`, `golangci-lint`, `Analyze (go)`, `Dependency Review`, `Scan for leaked secrets`.
- ✅ **Require conversation resolution before merging**.
- ✅ **Require linear history**.
- ✅ **Do not allow bypassing the above settings**.

---

## 4. 🛡️ Settings → Code security and analysis

Activa **todo lo que aparezca**:

- ✅ Dependency graph
- ✅ Dependabot alerts
- ✅ Dependabot security updates
- ✅ Dependabot version updates → ya tienes `.github/dependabot.yml`
- ✅ Code scanning → ya tienes CodeQL configurado
- ✅ Secret scanning + Push protection
- ✅ Private vulnerability reporting

---

## 5. 🔑 Settings → Secrets and variables → Actions

Crea estos **Repository secrets** para que funcione el deploy:

| Secret          | Valor                                                                 |
| --------------- | --------------------------------------------------------------------- |
| `VPS_HOST`      | IP o dominio de tu VPS (ej: `ski.midominio.com`)                      |
| `VPS_USER`      | Usuario SSH de despliegue (ej: `deploy`)                              |
| `VPS_PORT`      | Puerto SSH (`22` por defecto, omítelo si es 22)                       |
| `VPS_SSH_KEY`   | Clave privada **completa** OpenSSH (incluye `-----BEGIN...-----`)     |
| `VPS_APP_PATH`  | Ruta del proyecto en el VPS (ej: `/var/www/skihub`)                   |
| `VPS_SERVICE`   | Nombre del servicio systemd (ej: `skihub.service`)                    |
| `HEALTH_URL`    | URL pública de health (ej: `https://midominio.com/healthz`)           |

### Cómo generar la clave SSH de despliegue

```bash
# En tu máquina local o en el VPS
ssh-keygen -t ed25519 -f deploy_key -C "github-actions-deploy" -N ""

# Copia la PÚBLICA al VPS, al usuario deploy
ssh-copy-id -i deploy_key.pub deploy@TU_VPS

# Pega el contenido de la PRIVADA (deploy_key) en el secret VPS_SSH_KEY
cat deploy_key

# Borra la privada local después de pegarla
rm deploy_key deploy_key.pub
```

### Permisos sudo para reiniciar el servicio

En el VPS, como root:

```bash
echo 'deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart skihub.service, /bin/systemctl is-active skihub.service, /bin/systemctl status skihub.service' \
  | sudo tee /etc/sudoers.d/deploy-skihub
sudo chmod 440 /etc/sudoers.d/deploy-skihub
```

---

## 6. 🌍 Settings → Environments

Crea un environment llamado **`production`**:

- ✅ **Required reviewers** (te a ti mismo): el deploy esperará tu OK antes de tocar el VPS.
- ✅ **Deployment branches** → Selected branches → `main` únicamente.

Esto convierte cada deploy en una acción explícita y deja un registro auditado.

---

## 7. 📋 Crear el GitHub Project

1. Ve a la pestaña **Projects** del repo → **New project** → **Board**.
2. Llámalo `SkiHub Roadmap`.
3. Columnas sugeridas: `Backlog`, `Ready`, `In progress`, `In review`, `Done`.
4. **Workflows** del Project (engranaje arriba a la derecha):
   - "Item closed" → mover a `Done`.
   - "Pull request merged" → cerrar item asociado.
   - "Item added" → estado `Backlog`.

Crea las primeras tarjetas a partir de `docs/wiki/Roadmap.md`.

---

## 8. 📚 Subir la Wiki

GitHub Wiki es un repo Git aparte. Una sola vez:

```bash
# Activa Wiki desde Settings → General → Features (ya lo hiciste en el paso 2)
# Inicialízala creando una página dummy en la UI: pestaña Wiki → Create the first page → Save.

# Clónala
git clone https://github.com/celadaa/web_sky.wiki.git ../web_sky.wiki

# Sincroniza tus páginas
rsync -av --delete --exclude '.git' docs/wiki/ ../web_sky.wiki/

cd ../web_sky.wiki
git add .
git commit -m "docs: contenido inicial de la wiki"
git push
```

A partir de aquí, cada vez que actualices `docs/wiki/`, repite el `rsync` + commit.

---

## 9. 🏷️ Crear labels útiles

En Issues → Labels, crea los que faltan:

- `dependencias` (color gris) — usado por Dependabot.
- `triage` (rojo claro) — usado por las plantillas.
- `bug`, `enhancement`, `documentation`, `security`, `good first issue`,
  `help wanted` — los más estándar; GitHub trae algunos por defecto.

---

## 10. ✅ Verificación final

Después de hacer push y configurar todo:

```bash
# 1. Comprueba que CI está en verde
open https://github.com/celadaa/web_sky/actions

# 2. Mira la pestaña Security
open https://github.com/celadaa/web_sky/security

# 3. Mira los badges en el README — deberían estar verdes en cuestión de minutos
open https://github.com/celadaa/web_sky

# 4. Lanza un deploy a mano para probar
# Actions → Deploy to VPS → Run workflow → main
```

Si algo falla, revisa los logs en la pestaña Actions del workflow específico.

---

## 11. Bonus: lo siguiente que merece la pena

Cuando lo de arriba esté rodando, los siguientes saltos de calidad son:

1. **`/healthz` endpoint** que devuelva `200 OK` con JSON `{"status":"ok","version":"<sha>"}`.
   Lo usa el deploy y permite monitorización externa (UptimeRobot, BetterStack).
2. **Métricas Prometheus** (`/metrics`) para latencia y errores.
3. **Sentry** o equivalente para errores en producción.
4. **Backups diarios** de Postgres a almacenamiento externo (cron + `pg_dump` + `aws s3 cp`).
5. **Release Please** o `goreleaser` para automatizar versiones y changelog.
6. **Lighthouse CI** para vigilar performance/SEO de las páginas públicas.

---

Cualquier duda, abre un issue con la nueva plantilla 🏔️
