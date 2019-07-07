package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
)

var (
	client     *storage.Client
	bucketName string
)

func init() {
	var err error
	client, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("failed to new storage client: %v", err)
	}
	bucketName = os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("env BUCKET_NAME is empty")
	}
}

func checkQueryParam(values url.Values) (int, int, bool) {
	w := values.Get("w")
	h := values.Get("h")
	wi, err := strconv.Atoi(w)
	if err != nil {
		log.Printf("width param is not integer")
		return 0, 0, false
	}
	hi, err := strconv.Atoi(h)
	if err != nil {
		log.Printf("width param is not integer")
		return 0, 0, false
	}
	return wi, hi, true
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(name)
	attrs, err := obj.Attrs(r.Context())
	if err != nil {
		log.Printf("failed to get object attrs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentType := attrs.ContentType
	reader, err := obj.NewReader(r.Context())
	if err != nil {
		log.Printf("failed to new reader from object: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	width, height, ok := checkQueryParam(r.URL.Query())
	if !ok {
		w.Header().Add("Content-Type", contentType)
		w.Header().Add("Cache-Control", "public, max-age=604800, s-max-age=604800")
		if _, err := io.Copy(w, reader); err != nil {
			log.Printf("failed to writer response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	img, err := imaging.Decode(reader)
	if err != nil {
		log.Printf("failed to decode image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", contentType)
	w.Header().Add("Cache-Control", "public, max-age=604800, s-max-age=604800")
	img = imaging.Resize(img, width, height, imaging.Lanczos)
	switch contentType {
	case "image/jpeg", "image/jpg":
		err = imaging.Encode(w, img, imaging.JPEG)
	default:
		err = imaging.Encode(w, img, imaging.PNG)
	}
	if err != nil {
		log.Printf("failed to encode image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	http.HandleFunc("/", imageHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
