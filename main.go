package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	
	
	http.Handle("/", templ.Handler( html(home())))
	
	http.Handle("/intel", templ.Handler( html(withNavigation(intel()))))
	http.Handle("/drafts", templ.Handler( html(withNavigation(drafts()))))
	http.Handle("/signals", templ.Handler( html(withNavigation(signals()))))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}