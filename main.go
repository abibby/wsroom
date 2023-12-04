package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

func main() {

	log.Print("Listening on http://localhost:3339")
	http.ListenAndServe(":3339", JoinRoom())
}
