# SkiHub — Wiki

> Portal web para comparar estaciones de esquí, comprar forfaits y consultar pistas en directo.

Esta Wiki recoge documentación viva del proyecto que no encaja en el README: arquitectura,
operación, decisiones técnicas y runbook.

## Páginas

- 🏛 [Arquitectura](Arquitectura) — capas, flujo de petición, esquema general.
- 🚀 [Despliegue en VPS](Despliegue-en-VPS) — desde clonar hasta tener Nginx + systemd + HTTPS.
- 🔧 [Runbook operacional](Runbook-Operacional) — qué hacer cuando algo falla.
- 🗄 [Modelo de datos](Modelo-de-datos) — tablas, relaciones, migraciones.
- 🔐 [Seguridad](Seguridad) — cookies, contraseñas, headers, hardening.
- 🧪 [Testing](Testing) — convenciones de tests y cobertura.
- 📈 [Roadmap](Roadmap) — qué viene, qué está hecho.

## Enlaces útiles

- Repositorio: <https://github.com/celadaa/web_sky>
- Issues: <https://github.com/celadaa/web_sky/issues>
- Discussions: <https://github.com/celadaa/web_sky/discussions>
- Security policy: <https://github.com/celadaa/web_sky/security/policy>

## Convenciones

Las páginas de esta wiki:

- Se editan en `docs/wiki/` dentro del repo principal y se sincronizan
  manualmente al wiki de GitHub (mira [Despliegue de Wiki](#desplegar-cambios-a-la-wiki)).
- Usan `kebab-case` en los nombres de archivo (con guiones) para que los enlaces
  internos funcionen tanto en GitHub Wiki como visualizando los archivos en el repo.

### Desplegar cambios a la Wiki

```bash
# Una sola vez:
git clone https://github.com/celadaa/web_sky.wiki.git ../web_sky.wiki

# Cada vez que actualices docs/wiki/:
rsync -av --delete --exclude '.git' docs/wiki/ ../web_sky.wiki/
cd ../web_sky.wiki
git add . && git commit -m "docs: sync wiki" && git push
```
