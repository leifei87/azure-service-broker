package main

import (
	"net/http"
	"os"
	"fmt"
)

func main() {
	api := NewAppRouter()
	http.Handle("/", api)
	port := os.Getenv("PORT")
	fmt.Println(port)
	http.ListenAndServe(":"+port, nil)
}
