package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/xintamosaik/vc24/pages"
)

type IntelMeta struct {
	Description string    `json:"description"`
	Locked      bool      `json:"locked"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

const static_folder = "static"

func ssg(component templ.Component, filename string) {
	file, err := os.Create(static_folder + "/" + filename)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}

	err = html(component).Render(context.Background(), file)
	if err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

}

// this helper extracts all the alphabetic characters from the input string
func sanitizeFilename(input string) string {
	result := ""
	for _, char := range input {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			result += string(char)
		}
		// if there is a space, we can also add it
		if char == ' ' {
			result += " "
		}
		// if there is a dash, we can also add it
		if char == '-' {
			result += "_"
		}
		// if there is an underscore, we can also add it
		if char == '_' {
			result += "_"
		}
		// if there is a dot, we can also add it
		if char == '.' {
			result += "_"
		}
		// if there is a comma, we can also add it
		if char == ',' {
			result += "_"
		}
		// if there is a semicolon, we can also add it
		if char == ';' {
			result += "_"
		}
	}
	// If there are more than one subsequent underscores, replace them with a single underscore
	for i := 0; i < len(result)-1; i++ {
		if result[i] == '_' && result[i+1] == '_' {
			result = result[:i+1] + result[i+2:]
			i--
		}
	}
	// Trim leading and trailing underscores
	if len(result) > 0 && result[0] == '_' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == '_' {
		result = result[:len(result)-1]
	}
	// If the result is empty, return a default value
	if result == "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		return "default_" + timestamp

	}
	return result
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
func main() {

	ssg(withNavigation(pages.Home()), "index.html")

	// Serve dynamic pages

	http.Handle("/intel", templ.Handler(html(withNavigation(pages.Intel()))))
	//http.Handle("/intel/new", templ.Handler(html(withNavigation(newIntel()))))
	http.HandleFunc("/intel/new", handleIntelAdd)

	http.Handle("/drafts", templ.Handler(html(withNavigation(pages.Drafts()))))
	http.Handle("/signals", templ.Handler(html(withNavigation(pages.Signals()))))

	// Serve static files (SSG)
	http.Handle("/", http.FileServer(http.Dir(static_folder)))

	// Start the server
	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
