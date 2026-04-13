package main

import (
	"fmt"
	"log"
	"net/http"

	calc "codeagent-lab/calc"
)

func main() {
	http.HandleFunc("/api/window-summary", calc.HandleWindowSummary)

	addr := ":8080"
	fmt.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
