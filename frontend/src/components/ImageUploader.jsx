import React, { useState, useRef } from "react";

const IMAGE_LIMIT = import.meta.env.VITE_IMAGE_LIMIT ? parseInt(import.meta.env.VITE_IMAGE_LIMIT) : 16;
const MAX_TOTAL_SIZE_MB = import.meta.env.VITE_MAX_TOTAL_SIZE_MB ? parseInt(import.meta.env.VITE_MAX_TOTAL_SIZE_MB) : 100;
const BACKEND_URL = import.meta.env.VITE_STITCH_URL || "/api/stitch";

export default function ImageUploader() {
    const [images, setImages] = useState([]);
    const [uploading, setUploading] = useState(false);
    const [resultUrl, setResultUrl] = useState(null);
    const [error, setError] = useState("");
    const fileInputRef = useRef(null);

    const isDuplicate = (file, existingFiles) => {
        return existingFiles.some((f) => f.name === file.name && f.size === file.size);
    };

    const totalSizeMB = (files) => files.reduce((sum, f) => sum + f.size, 0) / (1024 * 1024);

    const validateFiles = (incomingFiles) => {
        const uniqueNewFiles = incomingFiles.filter((file) => !isDuplicate(file, images));

        if (uniqueNewFiles.length < incomingFiles.length) {
            setError("⚠️ Some duplicate images were skipped.");
        } else {
            setError("");
        }

        const newImages = [...images, ...uniqueNewFiles];
        const sizeMB = totalSizeMB(newImages);

        if (newImages.length > IMAGE_LIMIT) {
            setError(`⚠️ Max ${IMAGE_LIMIT} images allowed.`);
            return null;
        }
        if (sizeMB > MAX_TOTAL_SIZE_MB) {
            setError(`⚠️ Max total size is ${MAX_TOTAL_SIZE_MB} MB.`);
            return null;
        }

        return uniqueNewFiles;
    };

    const handleFileChange = (e) => {
        const files = Array.from(e.target.files);
        const newFiles = validateFiles(files);
        if (newFiles) {
            setImages((prev) => [...prev, ...newFiles]);
            setResultUrl(null);
        }
    };

    const handleDrop = (e) => {
        e.preventDefault();
        const files = Array.from(e.dataTransfer.files);
        const newFiles = validateFiles(files);
        if (newFiles) {
            setImages((prev) => [...prev, ...newFiles]);
            setResultUrl(null);
        }
    };

    const handleDragOver = (e) => {
        e.preventDefault();
    };

    const removeImage = (index) => {
        setImages((prev) => prev.filter((_, i) => i !== index));
        setError("");
    };

    const handleSubmit = async () => {
        if (images.length === 0) return;

        const formData = new FormData();
        images.forEach((img) => formData.append("images[]", img));

        setUploading(true);
        setResultUrl(null);
        setError("");

        try {
            const res = await fetch(BACKEND_URL, {
                method: "POST",
                body: formData,
            });

            if (!res.ok) throw new Error("Failed to stitch image");

            const blob = await res.blob();
            const url = URL.createObjectURL(blob);
            setResultUrl(url);
        } catch (err) {
            setError("Error: " + err.message);
        } finally {
            setUploading(false);
        }
    };

    return (
        <div className="max-w-xl mx-auto p-4 space-y-4">
            <h2 className="text-2xl font-bold">🧵 Upload Images for Stitching</h2>

            <div
                onDrop={handleDrop}
                onDragOver={handleDragOver}
                className="border-2 border-dashed border-blue-400 rounded-md p-4 text-center cursor-pointer hover:bg-blue-50"
                onClick={() => fileInputRef.current.click()}
            >
                <p className="text-sm text-gray-500">Drag and drop images here, or click to select</p>
                <p className="text-xs text-gray-400">(Max {IMAGE_LIMIT} images, {MAX_TOTAL_SIZE_MB} MB total)</p>
                <input
                    type="file"
                    accept="image/*"
                    multiple
                    ref={fileInputRef}
                    onChange={handleFileChange}
                    className="hidden"
                />
            </div>

            {error && <div className="text-red-500 text-sm font-medium">{error}</div>}

            {images.length > 0 && (
                <div className="grid grid-cols-3 gap-2">
                    {images.map((img, index) => (
                        <div key={index} className="relative group">
                            <img
                                src={URL.createObjectURL(img)}
                                alt="preview"
                                className="w-full h-24 object-cover rounded border"
                            />
                            <button
                                type="button"
                                onClick={() => removeImage(index)}
                                className="absolute top-1 right-1 bg-red-800/70 text-black rounded-full w-6 h-6 items-center justify-center text-xs font-bold cursor-pointer hidden group-hover:flex"
                                title="Remove"
                            >
                                🗑️
                            </button>
                        </div>
                    ))}
                </div>
            )}

            <button
                onClick={handleSubmit}
                disabled={uploading || images.length === 0}
                className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:opacity-50"
            >
                {uploading ? "Stitching..." : "Generate Panorama"}
            </button>

            {uploading && (
                <div className="w-full bg-gray-200 rounded-full h-3">
                    <div className="bg-blue-500 h-3 rounded-full animate-pulse" style={{ width: "100%" }}></div>
                </div>
            )}

            {resultUrl && (
                <div className="pt-4">
                    <h3 className="font-semibold mb-2">🖼️ Result:</h3>
                    <img src={resultUrl} alt="Stitched panorama" className="w-full border rounded shadow" />
                    <a
                        href={resultUrl}
                        download="panorama.webp"
                        className="block mt-2 text-blue-600 hover:underline"
                    >
                        Download Image
                    </a>
                </div>
            )}
        </div>
    );
}
