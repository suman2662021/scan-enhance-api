# Image Compressar API Flow

Welcome to the **Image Compressar** codebase! This document is designed for absolute beginners to understand exactly how this application works from start to finish.

## 1. The Entry Point (`src/cmd/api/main.go`)
Every Go application starts at the `main()` function. In our case, this lives in `src/cmd/api/main.go`.

**What happens here?**
- **Load Configuration:** It reads settings (like the port to run on or folder paths) using `config.Load()`.
- **Set Up Storage:** It creates a `LocalStore` (to save image files) and a `FileStore` (to save data *about* the images, called metadata).
- **Initialize Processor:** It sets up a `DocumentProcessor` that actually does the work of changing the images.
- **Start the Server:** Finally, it creates the API `Server` and starts listening for incoming web requests on the configured port.

## 2. API Routes (`src/internal/api/server.go`)
Once the server is running, it listens for HTTP requests. The `Routes()` function connects URLs to specific logic (handlers).

- `GET /health` -> `handleHealth` (Just checks if the server is alive)
- `POST /v1/images/process` -> `handleProcess` (Processes a single image)
- `POST /v1/images/process/batch` -> `handleBatchProcess` (Processes multiple images at once)
- `GET /v1/images/{id}` -> `handleGetImage` (Retrieves metadata for an image)

## 3. The Core Logic: Processing an Image
Let's trace what happens when someone calls the main API: `POST /v1/images/process`

### Step A: Receiving the Request (`handleProcess` in `server.go`)
1. **Validation:** Checks if the request is a `POST` method.
2. **File Extraction:** Reads the uploaded file from the `image` field in the form data.
3. **Reading Data:** Converts the uploaded file into raw bytes and detects its type (e.g., JPEG or PNG) using the `readUpload` function.
4. **Parsing Settings:** Reads any extra settings the user sent (like `targetSizeKB` or enhancement options) using `parseRequest()`.
5. **Delegating:** Passes all this data to the `processOne()` function to do the heavy lifting.

### Step B: The Processing Pipeline (`processOne` in `server.go`)
This function is the conductor of the orchestra. 

1. **Generate an ID:** Creates a unique timestamp-based ID for the image.
2. **Save Original:** Saves the raw, untouched uploaded image to the hard drive using `s.files.SaveOriginal()`.
3. **Process Image:** Hands the raw image data to the `Processor` (which lives in `src/internal/processor/processor.go`).
4. **Save Processed:** Once the processor returns the enhanced/compressed image, it saves the new image to the hard drive using `s.files.SaveProcessed()`.
5. **Save Metadata:** Creates a record containing all details (original size, new size, paths, time taken) and saves it to a JSON file via `s.metadata.Save()`.

### Step C: The Actual Image Manipulation (`Process` in `src/internal/processor/processor.go`)
This is where the image is physically changed.

1. **Decoding:** Converts the raw bytes back into an Image object that Go understands.
2. **Enhancement (`applyEnhancements`):** Depending on the settings requested, it applies various filters:
   - *Brightness/Contrast:* Makes the image lighter or punchier.
   - *Shadow Removal:* A custom loop that compares pixels to average row brightness to remove uneven shadows.
   - *Blur/Sharpen:* Removes noise and makes edges crisp.
   - *Grayscale/Black & White:* Strips color and optionally forces pixels to be purely black or white (great for scanned documents).
3. **Encoding & Compression (`encodeToTarget`):** 
   - It tries to save the image as a JPEG or PNG.
   - If the user asked for a specific file size (`targetSizeKB`), the program uses a clever loop: it tries saving the JPEG at 14 different quality levels (from 92 down to 40) and picks the one that gets closest to the target size without going over.

## 4. Storage (`src/internal/storage/local.go`)
Whenever `server.go` wants to save a file, it calls functions here.
- It organizes files neatly into folders based on the current date: `imagefiles/original/YYYY/MM/DD/` and `imagefiles/processed/YYYY/MM/DD/`.
- It returns both the absolute path on the server and the public URL so the user can download the image later.

## Summary Flowchart
User Uploads Image 
      ↓
`main.go` starts the server 
      ↓
`server.go` (`handleProcess`) receives the HTTP request and reads the file 
      ↓
`server.go` (`processOne`) saves the original image
      ↓
`processor.go` (`Process`) applies filters and compresses to the target size
      ↓
`server.go` (`processOne`) saves the newly processed image and its metadata
      ↓
User receives a JSON response with the success details and image URLs!
