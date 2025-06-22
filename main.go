package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"

	"github.com/a-h/templ"
	"github.com/xintamosaik/vc24/pages"

)

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
		return "default_"+ timestamp
		
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
		filename := sanitizeFilename(title) + ".txt"

		description := r.FormValue("description")
		content := r.FormValue("content")

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
