package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongodb *mongo.Client

// Todo is the app object on MongoDB
type Todo struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Text      string             `json:"text"`
	IsDone    bool               `json:"is_done"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("v0.0.1 is up and running")) })
	s := r.PathPrefix("/api").Subrouter()

	mongodb = connectMongoDB()
	s.HandleFunc("/todo", handleCreate).Methods(http.MethodPost)
	s.HandleFunc("/todo", handleRead).Methods(http.MethodGet)
	s.HandleFunc("/todo", handleUpdate).Methods(http.MethodPut)
	s.HandleFunc("/todo", handleDelete).Methods(http.MethodDelete)

	port := os.Getenv("PORT")
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func connectMongoDB() *mongo.Client {
	mongoURI := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")
	return client
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var todo Todo
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &todo)
	if err != nil {
		log.Fatal(err)
	}
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()

	collection := mongodb.Database("golang").Collection("collection")
	insertResult, err := collection.InsertOne(context.TODO(), todo)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(insertResult)
}

func handleRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var results []*Todo

	collection := mongodb.Database("golang").Collection("collection")
	findOptions := options.Find()
	findOptions.SetLimit(10)
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var todo Todo
		err := cur.Decode(&todo)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &todo)
	}

	if err = cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(results)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var todo Todo
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &todo)
	if err != nil {
		log.Fatal(err)
	}
	todo.UpdatedAt = time.Now()
	collection := mongodb.Database("golang").Collection("collection")
	updateResult := collection.FindOneAndUpdate(context.TODO(), bson.M{"_id": todo.ID}, bson.M{"$set": bson.M{"text": todo.Text, "isdone": todo.IsDone, "updatedat": todo.UpdatedAt}})
	json.NewEncoder(w).Encode(updateResult)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var todo Todo
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &todo)
	if err != nil {
		log.Fatal(err)
	}
	collection := mongodb.Database("golang").Collection("collection")
	deleteResult := collection.FindOneAndDelete(context.TODO(), bson.M{"_id": todo.ID})
	json.NewEncoder(w).Encode(deleteResult)
}
