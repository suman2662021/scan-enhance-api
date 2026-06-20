package domain

import "time"

type EnhancementOptions struct {
	Grayscale        bool   `json:"grayscale"`
	BlackAndWhite    bool   `json:"blackAndWhite"`
	ShadowRemoval    bool   `json:"shadowRemoval"`
	Denoise          bool   `json:"denoise"`
	Sharpen          bool   `json:"sharpen"`
	ContrastBoost    bool   `json:"contrastBoost"`
	BrightnessAdjust bool   `json:"brightnessAdjust"`
	Mode             string `json:"mode"`
}

type ProcessRequest struct {
	TargetSizeKB int                `json:"targetSizeKB"`
	OutputFormat string             `json:"outputFormat"`
	Enhancement  EnhancementOptions `json:"enhancement"`
}

type ImageMetadata struct {
	ID                  string             `json:"id"`
	OriginalFilename    string             `json:"originalFilename"`
	StoredOriginalName  string             `json:"storedOriginalName"`
	StoredProcessedName string             `json:"storedProcessedName"`
	MIMEType            string             `json:"mimeType"`
	Width               int                `json:"width"`
	Height              int                `json:"height"`
	OriginalPath        string             `json:"originalPath"`
	ProcessedPath       string             `json:"processedPath"`
	OriginalURL         string             `json:"originalUrl"`
	ProcessedURL        string             `json:"processedUrl"`
	OriginalSizeBytes   int64              `json:"originalSizeBytes"`
	ProcessedSizeBytes  int64              `json:"processedSizeBytes"`
	TargetSizeBytes     int64              `json:"targetSizeBytes"`
	CompressionRatio    float64            `json:"compressionRatio"`
	ProcessingTimeMS    int64              `json:"processingTimeMs"`
	SHA256              string             `json:"sha256"`
	Status              string             `json:"status"`
	OutputFormat        string             `json:"outputFormat"`
	Enhancement         EnhancementOptions `json:"enhancement"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt"`
}

type BatchProcessResponse struct {
	TotalRequested int                `json:"totalRequested"`
	TotalProcessed int                `json:"totalProcessed"`
	TotalFailed    int                `json:"totalFailed"`
	Results        []ImageMetadata    `json:"results"`
	Errors         []BatchImageError  `json:"errors"`
}

type BatchImageError struct {
	FileName string `json:"fileName"`
	Error    string `json:"error"`
}
