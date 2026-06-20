# Image Compressar

This repository contains a Go-based image processing API that:

- accepts multipart image uploads
- stores originals in `imagefiles/original/`
- stores processed files in `imagefiles/processed/`
- writes JSON metadata records in `metadata/`
- applies a document-scan style enhancement pipeline
- compresses output toward a target file size

## Current Phase

The following production components are intentionally deferred in this phase:

- Queue: `RabbitMQ`
- Cache / rate limiting / idempotency support: `Redis`
- Containerization: Docker multi-stage build
- Observability: `OpenTelemetry`, `Prometheus`, `Grafana`, `Loki/ELK`

The service currently runs as a single local API process with filesystem-backed storage and JSON metadata records.

## Run

### 1. Create your local environment file

```bash
cp .env.example .env
```

### 2. Export the environment variables

```bash
export $(grep -v '^#' .env | xargs)
```

### 3. Start the API server

```bash
go run ./src/cmd/api
```

By default, the API will run at [http://localhost:8080](http://localhost:8080).

## API

### `POST /v1/images/process`

Multipart form fields:

- `image`: file upload, required
- `targetSizeKB`: numeric, optional, default `500`
- `outputFormat`: `jpg` or `png`, optional
- `mode`: defaults to `document_scan`
- `grayscale`
- `blackAndWhite`
- `shadowRemoval`
- `denoise`
- `sharpen`
- `contrastBoost`
- `brightnessAdjust`

Example:

```bash
curl -X POST http://localhost:8080/v1/images/process \
  -F "image=@/path/to/input.jpg" \
  -F "targetSizeKB=200" \
  -F "outputFormat=jpg" \
  -F "blackAndWhite=false"
```

The response includes:

- processed image path and URL
- original image path and URL
- final file size
- compression ratio
- processing time
- image metadata

### `POST /v1/images/process/batch`

Processes multiple images in a single request using the same processing options.

Rules:

- use the `image` form field multiple times
- `image[]` is also supported for Postman and frontend array-style uploads
- supports up to `100` files per request by default
- each image is processed independently
- one failed image does not fail the full batch

Postman-ready example:

```bash
curl --location 'http://localhost:8080/v1/images/process/batch' \
--form 'image[]=@"/absolute/path/to/input-1.jpg"' \
--form 'image[]=@"/absolute/path/to/input-2.png"' \
--form 'image[]=@"/absolute/path/to/input-3.jpg"' \
--form 'targetSizeKB="200"' \
--form 'outputFormat="jpg"' \
--form 'blackAndWhite="false"'
```

Batch response includes:

- `totalRequested`
- `totalProcessed`
- `totalFailed`
- `results` for successful images
- `errors` for failed images

### `GET /v1/images/{id}`

Returns stored metadata for the processed asset.

Example:

```bash
curl http://localhost:8080/v1/images/img_20260620T120000.000000000
```

### `GET /health`

Simple health check endpoint.

Example:

```bash
curl http://localhost:8080/health
```

## Notes

- The current implementation supports `JPEG` and `PNG`.
- The processor is intentionally isolated behind an interface so `libvips` and `OpenCV` can replace or augment the built-in pure-Go pipeline later.
- The deferred infrastructure pieces can be added back once we move from local phase to scaled deployment.
