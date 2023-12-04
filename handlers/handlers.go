package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	MongoMainDB            = "main"
	MongoCollectionVillage = "village"
)

type Handlers struct {
	mongo *mongo.Client
	db    *mongo.Database
}

func New(mongo *mongo.Client) *Handlers {
	h := Handlers{mongo, nil}
	h = h.Init()

	return &h
}

func (h Handlers) Init() Handlers {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify that the connection was created successfully.
	err := h.mongo.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	opts := &options.DatabaseOptions{
		ReadConcern:  readconcern.Majority(),
		WriteConcern: writeconcern.Majority(),
	}

	h.db = h.mongo.Database(MongoMainDB, opts)
	return h
}

func (h Handlers) Add(w http.ResponseWriter, req *http.Request) {
	var village Village

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&village)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	village.ID = primitive.NewObjectID().Hex()

	_, err = h.db.Collection(MongoCollectionVillage).InsertOne(req.Context(), village)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(village)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h Handlers) List(w http.ResponseWriter, req *http.Request) {
	cur, err := h.db.Collection(MongoCollectionVillage).Find(req.Context(), bson.D{})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var villages []Village
	err = cur.All(req.Context(), &villages)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if villages == nil {
		villages = []Village{}
	}

	err = json.NewEncoder(w).Encode(villages)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h Handlers) Like(w http.ResponseWriter, req *http.Request) {
	id, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var village Village
	err = h.db.Collection(MongoCollectionVillage).FindOneAndUpdate(req.Context(),
		bson.M{"_id": string(id)},
		bson.D{{"$inc", bson.M{"likes": 1}}},
		options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After),
	).Decode(&village)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(village)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h Handlers) Feed(w http.ResponseWriter, req *http.Request) {
	cur, err := h.db.Collection(MongoCollectionVillage).Find(req.Context(), bson.D{}, options.Find().
		SetSort(bson.M{"likes": -1}).
		SetLimit(1))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var feed []Village
	err = cur.All(req.Context(), &feed)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if feed == nil {
		feed = []Village{}
	}

	err = json.NewEncoder(w).Encode(feed)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
