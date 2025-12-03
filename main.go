package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/genai"
)

var (
	aiClient  *genai.Client
	modelName = "imagen-4.0-generate-001"
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

	log.Printf("API listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleTextToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req TextToImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		http.Error(w, "missing prompt", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	imgBytes, err := generateSingleImage(ctx, req.Prompt)
	if err != nil {
		http.Error(w, "generation error", http.StatusInternalServerError)
		return
	}

	writeImagePNG(w, imgBytes)
}

func handleResize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req ResizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" {
		http.Error(w, "missing image", http.StatusBadRequest)
		return
	}
	if req.Scale != 2 && req.Scale != 4 {
		http.Error(w, "scale must be 2 or 4", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf("Resize this image by x%d preserving details.", req.Scale)

	ctx := r.Context()
	imgBytes, err := generateSingleImage(ctx, prompt)
	if err != nil {
		http.Error(w, "resize error", http.StatusInternalServerError)
		return
	}

	writeImagePNG(w, imgBytes)
}

func handleSketchToImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req SketchToImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" || req.Description == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf("Sketch to Image: interpret this sketch as '%s'.", req.Description)

	ctx := r.Context()
	imgBytes, err := generateSingleImage(ctx, prompt)
	if err != nil {
		http.Error(w, "sketch error", http.StatusInternalServerError)
		return
	}

	writeImagePNG(w, imgBytes)
}

func handleMagicEraser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req MagicEraserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ImageBase64 == "" {
		http.Error(w, "missing image", http.StatusBadRequest)
		return
	}

	if _, err := base64.StdEncoding.DecodeString(req.ImageBase64); err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}

	prompt := "Magic eraser: remove the pink masked area and reconstruct background."

	ctx := r.Context()
	imgBytes, err := generateSingleImage(ctx, prompt)
	if err != nil {
		http.Error(w, "eraser error", http.StatusInternalServerError)
		return
	}

	writeImagePNG(w, imgBytes)
}

func generateSingleImage(ctx context.Context, prompt string) ([]byte, error) {
	cfg := &genai.GenerateImagesConfig{
		NumberOfImages: 1,
	}
	resp, err := aiClient.Models.GenerateImages(ctx, modelName, prompt, cfg)
	if err != nil {
		return nil, err
	}
	if len(resp.GeneratedImages) == 0 || resp.GeneratedImages[0].Image == nil {
		return nil, fmt.Errorf("no image returned")
	}
	return resp.GeneratedImages[0].Image.ImageBytes, nil
}

func writeImagePNG(w http.ResponseWriter, img []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(img)
}
