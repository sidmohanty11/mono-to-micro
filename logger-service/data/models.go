package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Data      string    `json:"data" bson:"data"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func New(mongo *mongo.Client) Models {
	client = mongo
	return Models{
		LogEntry: LogEntry{},
	}
}

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), LogEntry{
		ID:        primitive.NewObjectID().Hex(),
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error Inserting:", err.Error())
		return err
	}
	return nil
}

func (l *LogEntry) FindAll() ([]*LogEntry, error) {
	collection := client.Database("logs").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)

	if err != nil {
		log.Println("Error Finding:", err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*LogEntry
	for cursor.Next(ctx) {
		var entry LogEntry
		err := cursor.Decode(&entry)
		if err != nil {
			log.Println("Error Decoding:", err.Error())
			return nil, err
		}
		results = append(results, &entry)
	}
	return results, nil
}

func (l *LogEntry) FindOne(id string) (*LogEntry, error) {
	collection := client.Database("logs").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error Finding:", err.Error())
		return nil, err
	}

	var entry LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&entry)
	if err != nil {
		log.Println("Error Finding:", err.Error())
		return nil, err
	}
	return &entry, nil
}

func (l *LogEntry) Drop() error {
	collection := client.Database("logs").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := collection.Drop(ctx)
	if err != nil {
		log.Println("Error Dropping:", err.Error())
		return err
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	collection := client.Database("logs").Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	docID, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		log.Println("Error Finding:", err.Error())
		return nil, err
	}

	update := bson.D{{
		"$set", bson.D{
			{"name", l.Name},
			{"data", l.Data},
			{"updated_at", time.Now()},
		},
	}}

	res, err := collection.UpdateOne(ctx, bson.M{"_id": docID}, update)
	if err != nil {
		log.Println("Error Updating:", err.Error())
		return nil, err
	}

	return res, nil
}
