package main

import (
	"context"
	"log"
	"net/http"
	"study-mongodb/handlers"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://user:pass@localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	h := handlers.New(client)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/village/add", h.Add)
	mux.HandleFunc("/api/village/list", h.List)
	mux.HandleFunc("/api/village/like", h.Like)
	mux.HandleFunc("/api/feed", h.Feed)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
