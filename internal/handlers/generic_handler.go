package handlers

import (
	"log"
)

// Generic error handler
func Handle(err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
