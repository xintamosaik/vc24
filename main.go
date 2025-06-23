package main

import (
	"context"
	"strings"

	"fmt"
	"log"
	"net/http"
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

func handleAnnotationAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Missing filename ", http.StatusBadRequest)
		return
	}

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

	annotation := r.FormValue("selected_text") // the selected text might span multiple containers

	if annotation == "" {
		http.Error(w, "Missing annotation", http.StatusBadRequest)
		return
	}
	if len(annotation) == 0 {
		http.Error(w, "Annotation is empty", http.StatusBadRequest)
		return
	}

	startedAtPosition := r.FormValue("started_at_position") // This is the position in the start container that the slection started
	if startedAtPosition == "" {
		http.Error(w, "Missing started_at_position", http.StatusBadRequest)
		return
	}

	endedAtPosition := r.FormValue("ended_at_position") // This is the position in the end container that the selection ended
	if endedAtPosition == "" {
		http.Error(w, "Missing ended_at_position", http.StatusBadRequest)
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

	// we compute the "search window"
	// start = containerStart + positionStart
	// end = containerEnd + positionEnd

	container_start_in_content_index := strings.Index(content, startedAt)
	if container_start_in_content_index == -1 {
		http.Error(w, "Container start not found in content", http.StatusBadRequest)
		return
	}
	log.Printf("Container start found at index: %d", container_start_in_content_index)

	container_end_in_content_index := strings.Index(content, endedAt)
	if container_end_in_content_index == -1 {
		http.Error(w, "Container end not found in content", http.StatusBadRequest)
		return
	}
	log.Printf("Container end found at index: %d", container_end_in_content_index)

	annotations_fields := strings.Fields(annotation)
	annotations_fields_first := annotations_fields[0]
	log.Printf("First annotation field: %s", annotations_fields_first)
	annotations_fields_last := annotations_fields[len(annotations_fields)-1]
	log.Printf("Last annotation field: %s", annotations_fields_last)

	first_annotation_in_container_start := strings.Index(content[container_start_in_content_index:], annotations_fields_first)
	if first_annotation_in_container_start == -1 {
		http.Error(w, "First annotation not found in container start", http.StatusBadRequest)
		return
	}
	log.Printf("First annotation found at index: %d", first_annotation_in_container_start)

	last_annotation_in_container_end := strings.Index(content[container_end_in_content_index:], annotations_fields_last)
	if last_annotation_in_container_end == -1 {	
		http.Error(w, "Last annotation not found in container end", http.StatusBadRequest)
		return
	}
	log.Printf("Last annotation found at index: %d", last_annotation_in_container_end)
	
	window_start := container_start_in_content_index + first_annotation_in_container_start
	window_end := container_end_in_content_index + last_annotation_in_container_end + len(annotations_fields_last)
	window := content[window_start:window_end]
	log.Printf("Search window: %s", window)

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
