package main

import (
	"flag"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Templates
var uploadTemplate = template.Must(template.ParseFiles("upload.html"))
var thumbnailTemplate = template.Must(template.ParseFiles("thumbnails.html"))

// Directories
var staticURL = "/static/"
var staticDir = "static/"
var mediaURL = "/media/"
var mediaDir = "media/"

// TODO errors are currently ignored
var curDir, _ = os.Getwd()

// Server Address
var address string

// Common attrs
var attrs = map[string]interface{}{
	"static_url": staticURL,
}

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		displayThumbnails(w, req)
		return
	}
	uploadTemplate.Execute(w, attrs)
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(""))
}

type Thumbnail struct {
	Quality int
	Size    string
	URL     string
}

func displayThumbnails(w http.ResponseWriter, req *http.Request) {
	// contentType := req.Header.Get("Content-Type")
	// log.Println("Content type:", contentType)

	// TODO If image is in the request body, use:
	// img, imgErr := jpeg.Decode(req.Body)
	rawFile, header, fileErr := req.FormFile("image")
	if fileErr != nil {
		io.WriteString(w, fmt.Sprintf("Error while reading: %s", fileErr))
		return
	}

	// Handle jpg, gif, and png
	img, typeString, imgErr := image.Decode(rawFile)
	log.Println("Decoded image type:", typeString)
	if imgErr != nil {
		io.WriteString(w, fmt.Sprintf("Error while decoding: %s", imgErr))
		return
	}

	// TODO How to get file size without re-encoding?
	// TODO Shrink the original image down to thumbnail size (at most 200px)

	// Create thumbnails for a range of qualities
	// TODO User option
	qualities := []int{100, 90, 75, 65, 50, 35, 25, 10}
	files := make([]*Thumbnail, len(qualities))

	// TODO Create the image directory if it does not exist
	for index, quality := range qualities {
		// TODO Errors are ignored
		filename := fmt.Sprintf("_%d.jpg", quality)
		fullPath := filepath.Join(curDir, mediaDir, filename)
		url := filepath.Join(mediaURL, filename)

		// TODO Never write the file to disk - simply base64 encode
		output, _ := os.Create(fullPath)
		defer output.Close()

		jpeg.Encode(output, img, &jpeg.Options{quality})
		files[index] = &Thumbnail{Quality: quality, URL: url}

		info, infoErr := output.Stat()
		if infoErr == nil {
			files[index].Size = strconv.FormatFloat(float64(info.Size())/1000.0, 'f', 1, 64) + " kb"
		}
	}

	// TODO Easiest way to extend another map
	templateAttrs := make(map[string]interface{})
	for key, value := range attrs {
		templateAttrs[key] = value
	}
	templateAttrs["Filename"] = header.Filename
	templateAttrs["Files"] = files

	thumbnailTemplate.Execute(w, templateAttrs)
}

func main() {
	// Parse the given server address
	flag.StringVar(&address, "address", ":8080", "address for the server")
	flag.Parse()

	// Static files: images, css, javascript
	http.Handle(staticURL, http.StripPrefix(staticURL, http.FileServer(http.Dir(filepath.Join(curDir, staticDir)))))

	// Uploaded files
	http.Handle(mediaURL, http.StripPrefix(mediaURL, http.FileServer(http.Dir(filepath.Join(curDir, mediaDir)))))

	http.HandleFunc("/", uploadHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	log.Printf("Server starting on address: %s\n", address)

	serveErr := http.ListenAndServe(address, nil)
	if serveErr != nil {
		log.Println(serveErr)
	}
}
