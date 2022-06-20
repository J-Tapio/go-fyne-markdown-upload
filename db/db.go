package db

import (
	"context"
	"time"
	"os"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/bson"
)

/*
Referencing
https://www.mongodb.com/blog/post/quick-start-golang-mongodb-starting-and-setup
*/

type NoteFile struct {
	Title string		`bson:"title, omitempty"`
	File []byte			`bson:"file, omitempty"`
	Tags []string		`bson:"tags, omitempty"`
}

type dbClient struct {
	client *mongo.Client
	collection *mongo.Collection
}

var db dbClient

func Initialize() {
	// Load local environment file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error while loading .env file")
	}

	// Connect to Database
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("ATLAS_URI")))

	if err != nil {
		log.Fatalf("Error while connecting to Mongo cluster, error:\n%s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	defer cancel()
	db.client = client
	db.collection = client.Database("markdown-notes").Collection("notes")
	log.Println("Connected to database")
}

func CheckForDuplicateNote(title string) (int64, error) {
	log.Println("Searching for duplicate titles from collection")
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	noteCount, err := db.collection.CountDocuments(ctx, bson.M{"title": title})

	if err != nil {
		return 0, err
	} else {
		return noteCount, nil
	}
}

func UploadNote(file NoteFile) (bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	document := bson.M{"title": file.Title, "tags": file.Tags, "file": file.File}
	result, err := db.collection.InsertOne(ctx, document)
	if err != nil {
		log.Println(err)
		return false
	} else {
		log.Printf("Inserted document with _id: %v\n", result.InsertedID)
		return true
	}
}

func CloseDb() {
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	log.Println("Closing database connection")
	if err := db.client.Disconnect(ctx); err !=nil {
		log.Fatalf("Issue with closing database connection: %s\n", err)
	}
}
