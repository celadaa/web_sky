# Modelos 3D — Snowbreak Market

Carpeta reservada para los modelos GLB que se cargarán desde Three.js
en la siguiente fase. La landing premium (`/premium`) ya está
preparada para sustituir los placeholders 2D por canvas 3D sin tocar
el HTML.

## Modelos esperados

| Archivo            | Producto                | data-product-id |
|--------------------|-------------------------|-----------------|
| `ski-basic.glb`    | Pack Esquí Básico       | `ski-basic`     |
| `ski-premium.glb`  | Pack Esquí Premium      | `ski-premium`   |
| `snowboard.glb`    | Pack Snowboard          | `snowboard`     |
| `kids-pack.glb`    | Pack Infantil           | `kids-pack`     |
| `helmet.glb`       | Casco + Seguridad       | `helmet`        |
| `insurance.glb`    | Seguro de nieve         | `insurance`     |
| `wax.glb`          | Mantenimiento / encerado| `wax`           |

## Convenciones recomendadas

- **Eje Y arriba**, escala normalizada a ~1 unidad de alto.
- **Centro del modelo en el origen** (Blender: Object → Set Origin → Geometry to Origin).
- **Material PBR** con metallic/roughness. Mantén la paleta monocroma del proyecto.
- Exporta desde Blender con: glTF Binary (.glb), Selected Objects, +Y Up.

## Cómo enchufar Three.js cuando lleguen los modelos

1. Añadir a `premium.tmpl` (antes de `animations.js`):
   ```html
   <script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r155/three.min.js"></script>
   <script src="/static/js/three-loader.js"></script>
   ```
   (Three.js está permitido por la CSP — viene desde cdnjs.)

2. Crear `web/static/js/three-loader.js` que use `GLTFLoader` para cada
   `[data-product-viz]` con `data-model` no vacío, monte un canvas en
   ese div y deje el placeholder oculto.

3. El selector ya está expuesto en `window.SBMarket.products` para
   iterar productos sin tocar el DOM.

Mientras no haya `.glb`, el placeholder 2D con SVG, grid y stamp
"SB · 3D READY" sirve perfectamente para portfolio.
