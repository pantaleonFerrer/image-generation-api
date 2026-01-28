# API de Generación de Imágenes




:
API REST desarrollada en Go que utiliza Google GenAI (Imagen 4.0) para generar y manipular imágenes mediante inteligencia artificial.

## 🚀 Características

- **Generación de imágenes desde texto**: Crea imágenes a partir de descripciones en lenguaje natural
- **Redimensionamiento inteligente**: Amplía imágenes manteniendo la calidad y los detalles
- **Conversión de bocetos**: Transforma dibujos o bocetos en imágenes realistas
- **Magic Eraser**: Elimina objetos o áreas específicas de imágenes y reconstruye el fondo

## 📋 Requisitos

- Go 1.25 o superior
- Docker y Docker Compose (opcional, para ejecutar en contenedor)
- API Key de Google Cloud Platform con acceso a GenAI

## ⚙️ Configuración

1. Crea un archivo `.env` en la raíz del proyecto:

```env
GOOGLE_API_KEY=tu_api_key_aqui
PORT=8080
```

2. Obtén tu API Key de Google Cloud Platform:
   - Ve a [Google Cloud Console](https://console.cloud.google.com/)
   - Habilita la API de GenAI
   - Crea credenciales y obtén tu API Key

## 🛠️ Instalación

### Opción 1: Ejecutar localmente

```bash
# Instalar dependencias
go mod download

# Ejecutar la aplicación
go run main.go
```

### Opción 2: Ejecutar con Docker

```bash
# Construir y ejecutar con Docker Compose
docker-compose up --build
```

La API estará disponible en `http://localhost:8080`

## 🔐 Autenticación

Todos los endpoints requieren una API Key válida. Se generan automáticamente **18 API Keys** al iniciar la aplicación, cada una con un límite de **20 llamadas**.

### Cómo usar la API Key

Puedes enviar la API Key de dos formas:

1. **Header HTTP** (recomendado):
   ```bash
   X-API-Key: tu_api_key_aqui
   ```

2. **Query parameter**:
   ```
   ?api_key=tu_api_key_aqui
   ```

### Obtener las API Keys

Las API Keys se generan automáticamente al iniciar la aplicación y se guardan en el archivo `api-keys.txt` en la raíz del proyecto.

También puedes consultar el estado de todas las keys mediante el endpoint:
```bash
GET /api-keys
```

Este endpoint devuelve todas las keys con su estado (usadas/limite restante).

### Límites

- Cada API Key tiene un límite de **20 llamadas**
- Una vez alcanzado el límite, recibirás un error `429 Too Many Requests`
- Las keys se reinician al reiniciar la aplicación

## 📚 Documentación de Endpoints

Todos los endpoints aceptan peticiones `POST` y devuelven imágenes en formato PNG. **Todos requieren autenticación mediante API Key.**

### 0. Listar API Keys

Obtiene el estado de todas las API keys disponibles.

**Endpoint:** `GET /api-keys`

**Respuesta:**
```json
{
  "keys": [
    {
      "key": "api_key_1",
      "used": 5,
      "limit": 20,
      "remaining": 15
    },
    ...
  ],
  "total": 18
}
```

---

### 1. Generar Imagen desde Texto

Genera una imagen a partir de una descripción en texto.

**Endpoint:** `POST /text-to-image`

**Request Body:**
```json
{
  "prompt": "Un gato naranja sentado en un jardín soleado"
}
```

**Parámetros:**
- `prompt` (string, requerido): Descripción de la imagen que deseas generar

**Respuesta:**
- **200 OK**: Imagen PNG generada
- **400 Bad Request**: Si falta el prompt o el body es inválido
- **500 Internal Server Error**: Error al generar la imagen

**Ejemplo con cURL:**
```bash
curl -X POST http://localhost:8080/text-to-image \
  -H "X-API-Key: tu_api_key_aqui" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Un paisaje montañoso al atardecer"}' \
  --output imagen.png
```

**Respuestas de error:**
- **401 Unauthorized**: Si falta la API key o es inválida
- **429 Too Many Requests**: Si se ha excedido el límite de llamadas de la API key

---

### 2. Redimensionar Imagen

Amplía una imagen manteniendo los detalles. Solo soporta escalado x2 o x4.

**Endpoint:** `POST /resize`

**Request Body:**
```json
{
  "image_base64": "iVBORw0KGgoAAAANSUhEUgAA...",
  "scale": 2
}
```

**Parámetros:**
- `image_base64` (string, requerido): Imagen codificada en Base64
- `scale` (int, requerido): Factor de escalado. Solo acepta valores `2` o `4`

**Respuesta:**
- **200 OK**: Imagen PNG redimensionada
- **400 Bad Request**: 
  - Si falta la imagen
  - Si el scale no es 2 o 4
  - Si el Base64 es inválido
- **500 Internal Server Error**: Error al redimensionar la imagen

**Ejemplo con cURL:**
```bash
# Primero convierte tu imagen a Base64
IMAGE_BASE64=$(base64 -i imagen_original.jpg)

curl -X POST http://localhost:8080/resize \
  -H "X-API-Key: tu_api_key_aqui" \
  -H "Content-Type: application/json" \
  -d "{\"image_base64\": \"$IMAGE_BASE64\", \"scale\": 2}" \
  --output imagen_redimensionada.png
```

---

### 3. Convertir Boceto a Imagen

Transforma un boceto o dibujo en una imagen realista basada en una descripción.

**Endpoint:** `POST /sketch-to-image`

**Request Body:**
```json
{
  "image_base64": "iVBORw0KGgoAAAANSUhEUgAA...",
  "description": "Una casa moderna con jardín"
}
```

**Parámetros:**
- `image_base64` (string, requerido): Boceto o dibujo codificado en Base64
- `description` (string, requerido): Descripción de cómo interpretar el boceto

**Respuesta:**
- **200 OK**: Imagen PNG generada a partir del boceto
- **400 Bad Request**: 
  - Si faltan campos requeridos
  - Si el Base64 es inválido
- **500 Internal Server Error**: Error al procesar el boceto

**Ejemplo con cURL:**
```bash
# Convierte tu boceto a Base64
SKETCH_BASE64=$(base64 -i boceto.jpg)

curl -X POST http://localhost:8080/sketch-to-image \
  -H "X-API-Key: tu_api_key_aqui" \
  -H "Content-Type: application/json" \
  -d "{\"image_base64\": \"$SKETCH_BASE64\", \"description\": \"Un coche deportivo rojo\"}" \
  --output imagen_final.png
```

---

### 4. Magic Eraser

Elimina áreas enmascaradas en color rosa de una imagen y reconstruye el fondo de forma inteligente.

**Endpoint:** `POST /magic-eraser`

**Request Body:**
```json
{
  "image_base64": "iVBORw0KGgoAAAANSUhEUgAA..."
}
```

**Parámetros:**
- `image_base64` (string, requerido): Imagen con áreas enmascaradas en rosa codificada en Base64

**Nota:** La imagen debe tener las áreas que deseas eliminar marcadas en color rosa (#FF00FF o similar).

**Respuesta:**
- **200 OK**: Imagen PNG con las áreas eliminadas y fondo reconstruido
- **400 Bad Request**: 
  - Si falta la imagen
  - Si el Base64 es inválido
- **500 Internal Server Error**: Error al procesar la imagen

**Ejemplo con cURL:**
```bash
# Convierte tu imagen con máscara rosa a Base64
IMAGE_BASE64=$(base64 -i imagen_con_mascara.jpg)

curl -X POST http://localhost:8080/magic-eraser \
  -H "X-API-Key: tu_api_key_aqui" \
  -H "Content-Type: application/json" \
  -d "{\"image_base64\": \"$IMAGE_BASE64\"}" \
  --output imagen_limpia.png
```

---

## 🔧 Variables de Entorno

| Variable | Descripción | Requerido | Valor por defecto |
|----------|-------------|-----------|-------------------|
| `GOOGLE_API_KEY` | API Key de Google Cloud Platform | Sí | - |
| `PORT` | Puerto en el que escucha la API | No | 8080 |

## 🐳 Docker

El proyecto incluye configuración de Docker Compose para facilitar el despliegue:

```bash
# Construir y ejecutar
docker-compose up --build

# Ejecutar en segundo plano
docker-compose up -d

# Ver logs
docker-compose logs -f

# Detener
docker-compose down
```

## 📝 Notas

- Todas las imágenes se devuelven en formato PNG
- El modelo utilizado es `imagen-4.0-generate-001` de Google GenAI
- Las imágenes en Base64 deben incluir el prefijo del tipo MIME si es necesario
- El endpoint de redimensionamiento solo acepta factores de escala 2x o 4x
- Para el Magic Eraser, las áreas a eliminar deben estar marcadas en color rosa en la imagen original

## 🐛 Manejo de Errores

La API devuelve códigos de estado HTTP estándar:

- **200 OK**: Operación exitosa
- **400 Bad Request**: Error en los parámetros de la petición
- **405 Method Not Allowed**: Método HTTP no permitido (solo POST)
- **500 Internal Server Error**: Error interno del servidor o de la API de Google

