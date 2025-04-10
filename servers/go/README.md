
# Pano Stitcher Demos

This repository contains a **Go Proxy** that demonstrates how to integrate the **Pano Stitcher API** into a Go project. It showcases both **HTTP** and **gRPC** integrations, allowing you to see how the Pano Stitcher API can be used for image stitching with various backend setups.

The purpose of this demo is to help developers understand how to use the **Pano Stitcher API** with Go projects and integrate it with a frontend.

---

## 🎯 **Purpose**

This repository demonstrates how to integrate the **Pano Stitcher API** in **Go projects**. It is designed to show how you can use the API to upload images, process them using Pano Stitcher, and get back a stitched panorama.

---

## 🛠️ **Prerequisites**

Before running this project, ensure you have the following:

### 1. **Go (1.16+)**

Install Go from the official website: https://golang.org/doc/install

### 2. **gRPC and Protocol Buffers**

You need to install `protoc` and `protoc-gen-go` to compile `.proto` files into Go.

```bash
brew install protobuf
go get google.golang.org/grpc
go get github.com/golang/protobuf/protoc-gen-go
```

### 3. **Pano Stitcher FastAPI gRPC server**

Ensure that the Pano Stitcher FastAPI service with gRPC support is running. This server is responsible for handling the image stitching requests.

---

## 🚀 **Getting Started**

### Step 1: Clone this Repository

```bash
git clone https://github.com/your-username/pano-stitcher-demos.git
cd pano-stitcher-demos/servers/go
```

---

### Step 2: Install Dependencies

Ensure that Go modules are set up and dependencies are installed:

```bash
go mod tidy
```

---

### Step 3: Prepare the `.env` File

Create a `.env` file in the root directory (`servers/go/`) to define environment variables.

```env
GRPC=true
PANO_KEY=your-secret-key
```

- **GRPC=true**: Set this to `true` to use gRPC for communication with the Pano Stitcher API. Otherwise, the proxy will default to HTTP.
- **PANO_KEY**: Provide your API key for authentication.

---

### Step 4: Build & Run the Go Proxy

To start the Go proxy server, run the following command:

```bash
go run main.go
```

- By default, the proxy will run on port `8080` for HTTP.
- If **GRPC=true** is set, the server will forward requests via gRPC.

---

## 🖥️ **How It Works**

### 1. **HTTP Integration (Fallback)**
When `GRPC=false` in the `.env` file, the proxy forwards requests to the Pano Stitcher API via HTTP.

### 2. **gRPC Integration**
When `GRPC=true` is set in the `.env` file, the proxy uses gRPC to send image data to the Pano Stitcher API for processing.

---

## 👨‍💻 **Integrating Pano Stitcher in Other Go Projects**

To integrate the Pano Stitcher API into your own Go project, follow the steps below.

### 1. **Install Dependencies**

Make sure your Go project has the necessary dependencies to communicate with the Pano Stitcher API:

```bash
go get google.golang.org/grpc
go get github.com/golang/protobuf/protoc-gen-go
```

### 2. **Generate Protobuf Files**

Ensure the `.proto` file used for gRPC communication is generated. Here’s how you can generate it:

```bash
protoc --proto_path=./proto --go_out=. --go-grpc_out=. ./proto/stitcher.proto
```

### 3. **Make the gRPC Call**

In your Go code, use the generated `stitcherpb` package and create a client for the Pano Stitcher service:

```go
conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

client := stitcherpb.NewStitcherClient(conn)

images := []*stitcherpb.ImageData{
    // Prepare image data here
}

resp, err := client.Process(context.Background(), &stitcherpb.StitchRequest{
    Images: images,
    Format: "webp",
    Key:    "your-api-key",
})
if err != nil {
    log.Fatalf("Failed to process images: %v", err)
}

fmt.Printf("Received stitched image: %v", resp.Filename)
```

---

### 4. **Configure Your API Key**

In the `.env` file, set your **Pano Stitcher API key**:

```env
PANO_KEY=your-api-key
```

---

### 5. **Call the Stitching API**

To use the API in your Go project, you can now call the proxy server (running HTTP or gRPC) to stitch images. For HTTP, send a `POST` request to `/stitch` with the images attached. For gRPC, the Go client sends a request directly to the Pano Stitcher server.

---

## 📝 **Troubleshooting**

- **HTTP Mode**: Ensure that the Pano Stitcher API (FastAPI or another backend) is correctly configured and running if you're using HTTP.
- **gRPC Mode**: Ensure that the gRPC server is running at `localhost:50051` (or the address configured in your `.env` file).
- **Missing Environment Variables**: Check that your `.env` file is properly loaded, and the keys are correctly set.

