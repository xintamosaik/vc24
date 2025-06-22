package main

import (
	"context"

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

func main() {

	ssg(withNavigation(pages.Home()), "index.html")

	// Serve dynamic pages

	http.HandleFunc("/intel", showIntelList)
	
	http.HandleFunc("/intel/new", handleIntelAdd)

	http.Handle("/drafts", templ.Handler(html(withNavigation(pages.Drafts()))))
	http.Handle("/signals", templ.Handler(html(withNavigation(pages.Signals()))))

	// Serve static files (SSG)
	http.Handle("/", http.FileServer(http.Dir(static_folder)))

	// Start the server
	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
