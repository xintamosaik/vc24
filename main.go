package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/a-h/templ"
	
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
func main() {
	// Build the pages

	// ssg index
	ssg(withNavigation(home()), "index.html")
	// ssg intel
	ssg(withNavigation(intel()), "intel.html")
	// ssg signals
	ssg(withNavigation(signals()), "signals.html")
	// ssg drafts
	ssg(withNavigation(drafts()), "drafts.html")
	
	
	// Generate static files (SSG)

	// Serve static files (SSG)
	http.Handle("/", http.FileServer(http.Dir(static_folder)))

	// Start the server
	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
