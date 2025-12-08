package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

var (
	aiClient  *genai.Client
	modelName = "gemini-3-pro-image-preview"
)

type TextToImageRequest struct {
	Prompt string `json:"prompt"`
}

type ResizeRequest struct {
	ImageBase64 string `json:"image_base64"`
	Scale       int    `json:"scale"`
}

type SketchToImageRequest struct {
	ImageBase64 string `json:"image_base64"`
	Description string `json:"description"`
}

type MagicEraserRequest struct {
	ImageBase64 string `json:"image_base64"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No se pudo cargar el archivo .env: %v", err)
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GOOGLE_API_KEY no está configurada. Por favor, configura la variable de entorno o crea un archivo .env")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("client error: %v", err)
	}

	aiClient = client

	http.HandleFunc("/text-to-image", handleTextToImage)
	http.HandleFunc("/resize", handleResize)
	http.HandleFunc("/sketch-to-image", handleSketchToImage)
	http.HandleFunc("/magic-eraser", handleMagicEraser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configurar servidor con límite de body aumentado (100MB por defecto)
	maxBodySize := int64(100 * 1024 * 1024) // 100MB
	if maxBodySizeEnv := os.Getenv("MAX_BODY_SIZE_MB"); maxBodySizeEnv != "" {
		var mb int64
		if _, err := fmt.Sscanf(maxBodySizeEnv, "%d", &mb); err == nil {
			maxBodySize = mb * 1024 * 1024
		}
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        http.MaxBytesHandler(http.DefaultServeMux, maxBodySize),
		ReadTimeout:    30 * 60 * 1000000000, // 30 minutos
		WriteTimeout:   30 * 60 * 1000000000, // 30 minutos
		MaxHeaderBytes: 1 << 20,              // 1MB para headers
	}

	log.Printf("API listening on :%s (Max body size: %d MB)", port, maxBodySize/(1024*1024))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleTextToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req TextToImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		writeError(w, "missing prompt", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	imgBytes, mimeType, err := generateSingleImage(ctx, req.Prompt)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		writeError(w, fmt.Sprintf("generation error: %v", err), http.StatusInternalServerError)
		return
	}

	writeImage(w, imgBytes, mimeType)
}

func handleResize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req ResizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" {
		writeError(w, "missing image", http.StatusBadRequest)
		return
	}
	if req.Scale != 2 && req.Scale != 4 {
		writeError(w, "scale must be 2 or 4", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		writeError(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf("Resize this image by x%d preserving details.", req.Scale)

	ctx := r.Context()
	imgBytes, mimeType, err := generateSingleImage(ctx, prompt)
	if err != nil {
		log.Printf("Error resizing image: %v", err)
		writeError(w, fmt.Sprintf("resize error: %v", err), http.StatusInternalServerError)
		return
	}

	writeImage(w, imgBytes, mimeType)
}

func handleSketchToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req SketchToImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" || req.Description == "" {
		writeError(w, "missing fields", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		writeError(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf("Interpret this sketch as '%s'.", req.Description)

	ctx := r.Context()
	imgBytes, mimeType, err := generateSingleImage(ctx, prompt)
	if err != nil {
		log.Printf("Error converting sketch to image: %v", err)
		writeError(w, fmt.Sprintf("sketch error: %v", err), http.StatusInternalServerError)
		return
	}

	writeImage(w, imgBytes, mimeType)
}

func handleMagicEraser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req MagicEraserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" {
		writeError(w, "missing image", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		writeError(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := "Remove the pink masked area and reconstruct the background."

	ctx := r.Context()
	imgBytes, mimeType, err := generateSingleImage(ctx, prompt)
	if err != nil {
		log.Printf("Error with magic eraser: %v", err)
		writeError(w, fmt.Sprintf("eraser error: %v", err), http.StatusInternalServerError)
		return
	}

	writeImage(w, imgBytes, mimeType)
}

func generateSingleImage(ctx context.Context, prompt string) ([]byte, string, error) {
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{
			"IMAGE",
			"TEXT",
		},
		ImageConfig: &genai.ImageConfig{
			ImageSize: "1K",
		},
	}

	for result, err := range aiClient.Models.GenerateContentStream(ctx, modelName, contents, config) {
		if err != nil {
			return nil, "", err
		}

		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
			continue
		}

		parts := result.Candidates[0].Content.Parts
		for _, part := range parts {
			if part.InlineData != nil {
				mimeType := part.InlineData.MIMEType
				if mimeType == "" {
					mimeType = "image/png"
				}
				return part.InlineData.Data, mimeType, nil
			}
		}
	}

	return nil, "", fmt.Errorf("no image returned")
}

func writeImage(w http.ResponseWriter, img []byte, mimeType string) {
	if mimeType == "" {
		mimeType = "image/png"
	}
	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)
	w.Write(img)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
