package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"pano-stitcher-go/proto/stitcherpb"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found or unable to load it")
	}

	if os.Getenv("GRPC") == "true" {
		fmt.Println("🚀 GRPC mode enabled. Forwarding HTTP requests to gRPC server...")
	} else {
		fmt.Println("🌐 HTTP mode enabled. Forwarding to pano stitcher via HTTP...")
	}

	http.HandleFunc("/stitch", handleStitch)
	port := "8080"
	fmt.Printf("🧵 Go Stitch Proxy running at http://localhost:%s/stitch\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// gRPC client stub to forward images to the Pano Stitcher API
func forwardViaGRPC(files []*multipart.FileHeader, w http.ResponseWriter) {
	fmt.Println("🚀 [gRPC] Connecting to pano stitcher gRPC service...")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := stitcherpb.NewStitcherClient(conn)
	fmt.Println("✅ [gRPC] Connected successfully")

	var images []*stitcherpb.ImageData
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		fmt.Println("📥 Adding file to gRPC request:", fh.Filename)
		images = append(images, &stitcherpb.ImageData{
			Filename: fh.Filename,
			Content:  buf.Bytes(),
		})
	}

	fmt.Println("📤 Sending gRPC request with", len(images), "image(s)")
	resp, err := client.Process(context.Background(), &stitcherpb.StitchRequest{
		Images: images,
		Format: "webp",
		Key:    os.Getenv("PANO_KEY"),
	})
	if err != nil {
		http.Error(w, "gRPC processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", resp.Filename))
	fmt.Println("🖼️ [gRPC] Received image with content-type:", resp.ContentType, "size:", len(resp.StitchedImage))
	w.WriteHeader(http.StatusOK)
	w.Write(resp.StitchedImage)
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

	// Parse uploaded files
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

	// If GRPC is enabled, use it instead of HTTP
	if os.Getenv("GRPC") == "true" {
		forwardViaGRPC(files, w)
		return
	}

	// HTTP fallback
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
