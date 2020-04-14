package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/firestore"
	"github.com/alexandrevicenzi/go-sse"
	"google.golang.org/api/option"
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
	s := sse.NewServer(&sse.Options{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
		},
	})
	defer s.Shutdown()

	// Configure the route
	// r := mux.NewRouter()
	// r.Use(mux.CORSMethodMiddleware(r))
	// r.Use(func(handler http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.Header().Set("Access-Control-Allow-Origin", "*")
	// 		w.Header().Set("Access-Control-Allow-Methods", "*")
	// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 		w.Header().Set("X-XSS-Protection", "1; mode=block")
	// 		w.Header().Set("X-Frame-Options", "DENY")
	// 		w.Header().Set("Strict-Transport-Security", "max-age=604800; includeSubDomains; preload")
	// 		handler.ServeHTTP(w, r)
	// 	})
	// })
	// r.Handle("/events", s)
	http.Handle("/events", s)

	chMsg := make(chan map[string]interface{})

	go func() {
		for {
			m := <-chMsg

			b, err := json.Marshal(&m)
			if err != nil {
				log.Println(err)
				continue
			}

			s.SendMessage("", sse.SimpleMessage(string(b)))
		}
	}()

	go watching(chMsg)

	log.Println("Listening at :3000")
	http.ListenAndServe(":3000", nil)
}

func watching(chMsg chan map[string]interface{}) {
	var fsClient *firestore.Client
	ctx := context.Background()
	var err error
	{
		opt := option.WithCredentialsFile("configs/firebase-credentials.json")
		fsClient, err = firestore.NewClient(ctx, projectID, opt)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	}
	defer fsClient.Close()

	iter := fsClient.Doc("patientData/FQdMC7APoo5wxL4GBqlm").Snapshots(ctx)
	defer iter.Stop()
	for {
		docsnap, err := iter.Next()
		if err != nil {
			log.Println("error", err)
		}
		fmt.Println(docsnap.Data())
		chMsg <- docsnap.Data()
	}
}
