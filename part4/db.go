package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"
	"log"
	"time"
)

func getClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Println(err)
	}
	return client
}

type message struct {
	Author  string
	Message string
	Time    string
}

func addChatMessage(client *mongo.Client, msg message) {
	mes := bson.M{"author": msg.Author, "message": msg.Message, "time": msg.Time}
	collection := client.Database("gocourse").Collection("chat")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, mes)
	if err != nil {
		log.Println(err)
	}
}
func getRecentMessages(client *mongo.Client, n int64) []message {
	collection := client.Database("gocourse").Collection("chat")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	findOptions := options.Find()
	findOptions.SetLimit(n)
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	cur, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		log.Println(err)
	}
	defer cur.Close(ctx)
	var s []message
	for cur.Next(ctx) {
		var result message
		err := cur.Decode(&result)
		if err != nil {
			log.Println(err)
		}
		s = append(s, result)
	}
	if err := cur.Err(); err != nil {
		log.Println(err)
	}
	return s
}
