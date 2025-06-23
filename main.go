package main

import (
	"context"

	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

func findWordInArray(word string, array []string) int {
	for i, w := range array {
		if w == word {
			return i // Return the index of the first occurrence
		}
	}
	return -1 // Return -1 if the word is not found
}

// the function takes two arrays of strings that has words and compares them.
// It tries to find the first occurence of the first word and then compares all the consecutive words
// If the next word is not a match it tries again until it can't find the first word in the second array
func compareWords(first, second []string) bool {
    if len(first) == 0 || len(second) == 0 {
        return false // Handle empty inputs gracefully
    }

    firstWord := first[0]
	firstIndex := findWordInArray(firstWord, second)
	if firstIndex == -1 {
		return false // First word not found in the second array
	}

	for i := 1; i < len(first); i++ {
		if first[i] != second[firstIndex+i] {
			log.Printf("Mismatch found: %s != %s at index %d", first[i], second[firstIndex+i], firstIndex+i)
			// If the next word is not a match, try to find the first word again
			firstIndex = findWordInArray(firstWord, second[firstIndex+1:])
			if firstIndex == -1 {
				return false // If we can't find the first word again, return false
			}
			firstIndex += 1 // Adjust the index to account for the slice
		}
	}
	return true // All words matched in order
    
}

func handleAnnotationAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// See if we find the file and if there is content in in

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Missing filename ", http.StatusBadRequest)
		return
	}
	file, err := os.Open("data/intel/" + filename)
	if err != nil {
		http.Error(w, "Failed to read intel file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	file_content, err := os.ReadFile("data/intel/" + filename)
	if err != nil {
		http.Error(w, "Failed to read intel file", http.StatusInternalServerError)
		return
	}

	content := string(file_content)
	if len(content) == 0 {
		http.Error(w, "Intel file is empty", http.StatusBadRequest)
		return
	}
	log.Printf("Received annotation for file: %s", filename)
	// We try to limit the area to search in
	startedAt := r.FormValue("started_at") // This is the container the selection started in
	endedAt := r.FormValue("ended_at")     // This is the container the selection ended in

	if startedAt == "" {
		http.Error(w, "Missing started_at", http.StatusBadRequest)
		return

	}
	if endedAt == "" {
		http.Error(w, "Missing ended_at", http.StatusBadRequest)
		return

	}

	// startedAtPosition := r.FormValue("started_at_position") // This is the position in the start container that the slection started
	// endedAtPosition := r.FormValue("ended_at_position") // This is the position in the end container that the selection ended


	log.Printf("Started at: %s, Ended at: %s", startedAt, endedAt)
	hasStart := strings.Contains(content, startedAt)
	if !hasStart {
		http.Error(w, "Started at not found in the content", http.StatusBadRequest)
		return
	}
	startedPosition := strings.Index(content, startedAt)
	if startedPosition == -1 {
		http.Error(w, "Started at not found in the content", http.StatusBadRequest)
		return
	}

	hasEnd := strings.Contains(content, endedAt)
	if !hasEnd {
		http.Error(w, "Ended at not found in the content", http.StatusBadRequest)
		return
	}
	endedPosition := strings.Index(content, endedAt) + len(endedAt) // + len(endedAt) to include the end container in the selection
	if endedPosition == -1 {
		http.Error(w, "Ended at not found in the content", http.StatusBadRequest)
		return
	}

	log.Printf("Started position: %d, Ended position: %d", startedPosition, endedPosition)
	
	annotation := r.FormValue("selected_text") // the selected text might span multiple containers
	
	if annotation == "" {
		http.Error(w, "Missing annotation", http.StatusBadRequest)
		return
	}
	if len(annotation) == 0 {
		http.Error(w, "Annotation is empty", http.StatusBadRequest)
		return
	}

	annotations_splitted_words := strings.Fields(annotation)
	if len(annotations_splitted_words) == 0 {
		http.Error(w, "Annotation is empty", http.StatusBadRequest)
		return
	}


	file_content_splitted := strings.Fields(content)
	if len(file_content_splitted) == 0 {
		http.Error(w, "Intel file is empty", http.StatusBadRequest)
		return
	}

	matched := compareWords(annotations_splitted_words, file_content_splitted)
	if !matched {
		http.Error(w, "Annotation does not match the content", http.StatusBadRequest)
		return
	}

	log.Printf("Annotation content: %s", annotation)
	keyword := r.FormValue("annotation")
	log.Println("Keyword for annotation:", keyword)

	html(withNavigation(pages.AnnotationSubmitted(filename))).Render(context.Background(), w)
}

func main() {

	ssg(withNavigation(pages.Home()), "index.html")

	// Serve dynamic pages

	http.HandleFunc("/intel", showIntelList)

	http.HandleFunc("/intel/new", handleIntelAdd)
	http.HandleFunc("/intel/{id}", handleIntelAnnotate)
	http.HandleFunc("/annotation/add", handleAnnotationAdd)

	http.Handle("/drafts", templ.Handler(html(withNavigation(pages.Drafts()))))
	http.Handle("/signals", templ.Handler(html(withNavigation(pages.Signals()))))

	// Serve static files (SSG)
	http.Handle("/", http.FileServer(http.Dir(static_folder)))

	// Start the server
	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
