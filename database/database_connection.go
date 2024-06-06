package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DbInstance() *mongo.Client{
	mongoDb := "mongodb://localhost:27017"
	fmt.Println(mongoDb)
	client,err := mongo.NewClient(options.Client().ApplyURI(mongoDb))
	if err != nil {
		log.Fatal(err)
	}
	cxt,cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	err = client.Connect(cxt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connection established successfully....")
	return client
}
var Client *mongo.Client = DbInstance()

// OpenCollection opens a collection in the specified database.
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    collection := client.Database("accounts").Collection(collectionName)
    return collection
}