package main

import (
	"fmt"
	"os"

	"github.com/KimHansenCubris/gh-godo/internal/server"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Workload Capacity UI running at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")

	if err := server.Start(addr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
