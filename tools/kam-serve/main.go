package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("/tmp/kam-root"))

	fmt.Println("Listening on port 9000 to serve directory contents /tmp/kam-root")

	err := http.ListenAndServe(":9000", fs)

	if err != nil {
		panic(err)
	}
}
