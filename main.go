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

func scoreMatch(annotation, window string) int {
	annotationWords := strings.Fields(annotation)
	windowWords := strings.Fields(window)

	score := 0
	for _, aw := range annotationWords {
		for _, ww := range windowWords {
			if aw == ww {
				score++
				break
			}
		}
	}
	return score
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
	window_fields := strings.Fields(window)
	window_fields_glued_back := strings.Join(window_fields, "")

	annotations_fields_glued_back := strings.Join(annotations_fields, "")

	good_enough_match := false
	if (window_fields_glued_back == annotations_fields_glued_back) {
		good_enough_match = true
		log.Println("Good enough match found")
	}

	matching_score := scoreMatch(annotation, window)
	log.Printf("Matching score: %d", matching_score)
	ratio := float64(matching_score) / float64(len(strings.Fields(annotation)))
	if ratio > 0.95 {
		good_enough_match = true
	}

	log.Printf("Matching ratio: %.2f", ratio)

	if !good_enough_match {
		http.Error(w, "Annotation does not match the content. Please notify the admin. His matching algorithm is dogwater.", http.StatusBadRequest)
		return
	}

	log.Printf("Annotation content: %s", annotation)
	keyword := r.FormValue("annotation")
	log.Println("Keyword for annotation:", keyword)

	// we persist the annotation. We will allow multiple annotations and even keywords. So we use a unix timestamp as the filename. There will never me more than one of them. 
	timestamp := fmt.Sprintf("%d", r.Context().Value("timestamp").(int64))
	annotationFile := "data/annotations/" + timestamp + ".json"
	// create the directory if it doesn't exist
	if err := os.MkdirAll("data/annotations", 0755); err != nil {
		http.Error(w, "Failed to create annotations directory", http.StatusInternalServerError)
		return
	}
	// create the file
	file, err = os.Create(annotationFile)
	if err != nil {
		http.Error(w, "Failed to create annotation file", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// write the annotation to the file

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
