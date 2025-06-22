package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"		

		"github.com/xintamosaik/vc24/pages"
)
type IntelMeta struct {
	Description string    `json:"description"`
	Locked      bool      `json:"locked"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

func handleIntelAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse the form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Process the form data here
		title := r.FormValue("title")
		sanitized := sanitizeFilename(title)
		filename := sanitized + ".txt"

		content := r.FormValue("content")
		intelPath := "data/intel/"
		// Ensure the directory exists
		if err := os.MkdirAll(intelPath, 0755); err != nil {
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}

		// Create or open the file
		file, err := os.Create(intelPath + filename)
		if err != nil {
			http.Error(w, "Failed to create file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		// Write the content to the file
		if _, err := file.WriteString(content); err != nil {
			http.Error(w, "Failed to write to file", http.StatusInternalServerError)
			return
		}

		metafile := sanitized + ".json"
		// Create or open the metadata file
		metaFile, err := os.Create(intelPath + metafile)
		if err != nil {
			http.Error(w, "Failed to create metadata file", http.StatusInternalServerError)
			return
		}
		defer metaFile.Close()

		description := r.FormValue("description")
		// Write metadata to the metadata file
		meta := IntelMeta{
			Description: description,
			Locked:      false,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		metaData, err := json.Marshal(meta)
		if err != nil {
			http.Error(w, "Failed to marshal metadata", http.StatusInternalServerError)
			return
		}
		// Write the metadata to the metadata file
		if _, err := metaFile.Write(metaData); err != nil {
			http.Error(w, "Failed to write metadata to file", http.StatusInternalServerError)
			return
		}

		// Log the received data (or handle it as needed)
		log.Printf("Received Intel: Title=%s, Description=%s, Content=%s", title, description, content)
		log.Printf("Sanitized Filename: %s", filename)
		// For example, you can read the form values and save them to a database or file
		html(withNavigation(pages.IntelSubmitted())).Render(context.Background(), w)
	} else {
		html(withNavigation(pages.IntelNew())).Render(context.Background(), w)
	}
}