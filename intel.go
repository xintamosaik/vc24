package main

import (
	"context"
	"encoding/json"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xintamosaik/vc24/pages"
)

type IntelMeta struct {
	Description string `json:"description"`
	Locked      bool   `json:"locked"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

const intelPath = "data/intel/"

func listIntelFiles() ([]string, error) {
	intelPath := "data/intel/"
	files, err := os.ReadDir(intelPath)
	if err != nil {
		return nil, err
	}

	var intelFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		filename := file.Name()
		dotIndex := strings.LastIndex(filename, ".")
		extension := ""
		if dotIndex == -1 {
			extension = ""
		} else {
			extension = filename[dotIndex+1:]
		}

		name := filename[:dotIndex]

		if extension == "json" {
			// Skip metadata files

			// But check if the corresponding intel file exists
			intelFile := intelPath + name + ".txt"
			if _, err := os.Stat(intelFile); os.IsNotExist(err) {
				log.Printf("Skipping metadata file %s because corresponding intel file does not exist", file.Name())
				continue // Skip this metadata file if the corresponding intel file does not exist
			}
			// If the intel file does not exist, skip this metadata file

		}

		if extension != "txt" && extension != "json" {
			log.Printf("Skipping file %s with unsupported extension %s", file.Name(), extension)
			continue // Skip files that are not .txt or .json
		}

		if extension == "txt" {
			// Check if the corresponding metadata file exists
			metaFile := intelPath + name + ".json"
			if _, err := os.Stat(metaFile); os.IsNotExist(err) {
				log.Printf("Skipping intel file %s because corresponding metadata file does not exist", file.Name())
				continue // Skip this intel file if the corresponding metadata file does not exist
			}
			intelFiles = append(intelFiles, file.Name())
		}

	}
	return intelFiles, nil
}

func showIntelList(w http.ResponseWriter, r *http.Request) {
	intelFiles, err := listIntelFiles()
	if err != nil {
		http.Error(w, "Failed to list intel files", http.StatusInternalServerError)
		return
	}

	// Render the intel list page with the list of files
	html(withNavigation(pages.Intel(intelFiles))).Render(context.Background(), w)
}
func handleIntelAnnotate(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Intel ID is required", http.StatusBadRequest)
		return
	}

	// Check if the file exists
	intelPath := "data/intel/" + id
	if _, err := os.Stat(intelPath); os.IsNotExist(err) {
		http.Error(w, "Intel file not found", http.StatusNotFound)
		return
	}

	// Read the content of the intel file
	file, err := os.ReadFile(intelPath)
	if err != nil {
		http.Error(w, "Failed to read intel file", http.StatusInternalServerError)
		return
	}
	content := string(file)

	splitted := strings.Split(content, "\n")
	for i, line := range splitted {
		splitted[i] = strings.TrimSpace(line)
	}

	// Render the annotation page for the specified intel file
	html(withNavigation(pages.IntelAnnotate(splitted, id))).Render(context.Background(), w)
}

func createIntelMeta(description string, filename string) error {
	metafile := filename + ".json"
	// Create or open the metadata file
	metaFile, err := os.Create(intelPath + metafile)
	if err != nil {
		return err
	}
	defer metaFile.Close()
	meta := IntelMeta{
		Description: description,
		Locked:      false,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	metaData, err := json.Marshal(meta)
	if err != nil {
		return err

	}
	// Write the metadata to the metadata file
	if _, err := metaFile.Write(metaData); err != nil {
		return err
	}

	return nil
}

func readIntelMeta(filename string) (*IntelMeta, error) {

	metaData, err := os.ReadFile(intelPath + filename)
	if err != nil {
		return nil, err
	}

	var meta IntelMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func writeIntelMeta(filename string, meta *IntelMeta) error {

	metaData, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	if err := os.WriteFile(intelPath+filename, metaData, 0644); err != nil {
		return err
	}

	return nil
}

func LockIntel(filename string) error {
	
	meta, err := readIntelMeta(filename)
	if err != nil {
		log.Printf("Error reading metadata for %s: %v", filename, err)
		return err
	}

	
	meta.Locked = true
	meta.UpdatedAt = time.Now().Unix()

	if err := writeIntelMeta(filename, meta); err != nil {
		log.Printf("Error writing metadata for %s: %v", filename, err)
		return err
	}

	return nil
}

func UnlockIntel(filename string) error {
	meta, err := readIntelMeta(filename)
	if err != nil {
		log.Printf("Error reading metadata for %s: %v", filename, err)
		return err
	}

	meta.Locked = false
	meta.UpdatedAt = time.Now().Unix()

	if err := writeIntelMeta(filename, meta); err != nil {
		log.Printf("Error writing metadata for %s: %v", filename, err)
		return err
	}

	return nil
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

		// write metadata to file
		if err := createIntelMeta(r.FormValue("description"), filename); err != nil {
			http.Error(w, "Failed to create metadata file", http.StatusInternalServerError)
			return
		}

		// Log the received data (or handle it as needed)
		log.Printf("Received Intel: Title=%s, Content=%s", title, content)
		log.Printf("Sanitized Filename: %s", filename)
		// For example, you can read the form values and save them to a database or file
		html(withNavigation(pages.IntelSubmitted())).Render(context.Background(), w)
	} else {
		html(withNavigation(pages.IntelNew())).Render(context.Background(), w)
	}
}
