package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/alexandrevicenzi/go-sse"
)

var projectID string

func init() {
	log.SetFlags(0)

	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID, _ = metadata.ProjectID()
	}
	if projectID == "" {
		log.Println("Could not determine Google Cloud Project. Running without log correlation. For local use set the GOOGLE_CLOUD_PROJECT environment variable.")
	}
}

func main() {
	// Create SSE server
	s := sse.NewServer(nil)
	defer s.Shutdown()

	// Configure the route
	http.Handle("/events", s)

	chMsg := make(chan string)
	
	// Send messages every 5 seconds
	go func() {
		for {
			s.SendMessage("/events/patient", sse.SimpleMessage(<-chMsg)
		}
	}()

	go watching(chMsg)

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}

func watching(chMsg chan string) {
	var fsClient *firestore.Client
	var err error
	{
		opt := option.WithCredentialsFile("configs/firebase-credentials.json")
		ctx := context.Background()
		fsClient, err = firestore.NewClient(ctx, projectID, opt)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	}
	defer fsClient.Close()

	iter := fsClient.Doc("patientData/{ID}").Snapshots(ctx)
	defer iter.Stop()
	for {
		docsnap, err := iter.Next()
		if err != nil {
			log.Println("error", err)
		}
		_ = docsnap // TODO: Use DocumentSnapshot.
		chMsg <- "updated"
	}
}
