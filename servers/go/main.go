package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found or unable to load it")
	}
	http.HandleFunc("/stitch", handleStitch)
	port := "8080"
	fmt.Printf("🧵 Go Stitch Proxy running at http://localhost:%s/stitch\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleStitch(w http.ResponseWriter, r *http.Request) {
	// CORS support
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-internal-key")

	// Handle preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "No images field found", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		formFile, err := writer.CreateFormFile("images", fileHeader.Filename)
		if err != nil {
			http.Error(w, "Failed to write form file", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(formFile, file); err != nil {
			http.Error(w, "Failed to copy file contents", http.StatusInternalServerError)
			return
		}
	}

	writer.Close()

	fmt.Println("✅ Forwarding", len(files), "file(s)")
	fmt.Println("📦 Buffer size (bytes):", buf.Len())

	panoURL := os.Getenv("PANO_URL")
	if panoURL == "" {
		panoURL = "http://localhost:8000/stitch"
	}

	req, err := http.NewRequest("POST", panoURL, &buf)
	if err != nil {
		http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-internal-key", os.Getenv("PANO_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach pano stitcher", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
