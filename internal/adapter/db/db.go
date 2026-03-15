package db

import (
	"context"
	"time"

	"github.com/pkvartsianyi/geoint-harvester/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type GeoJSONPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type ScrapedMessage struct {
	Channel     string        `bson:"channel"`
	Content     string        `bson:"content"`
	MsgID       int           `bson:"msg_id"`
	Timestamp   time.Time     `bson:"timestamp"`
	Geolocation *GeoJSONPoint `bson:"geolocation,omitempty"`
}

type MongoAdapter struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoAdapter(ctx context.Context, uri string) (*MongoAdapter, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database("telegram_scraper").Collection("scraped_data")

	return &MongoAdapter{
		client:     client,
		collection: collection,
	}, nil
}

func (a *MongoAdapter) UpsertMessage(ctx context.Context, msg domain.Message) error {
	var geo *GeoJSONPoint
	if msg.Geolocation != nil {
		geo = &GeoJSONPoint{
			Type:        msg.Geolocation.Type,
			Coordinates: msg.Geolocation.Coordinates,
		}
	}

	dbMsg := ScrapedMessage{
		Channel:     msg.Channel,
		Content:     msg.Content,
		MsgID:       msg.MsgID,
		Timestamp:   msg.Timestamp,
		Geolocation: geo,
	}

	filter := bson.M{"channel": dbMsg.Channel, "msg_id": dbMsg.MsgID}
	update := bson.M{"$set": dbMsg}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := a.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (a *MongoAdapter) Exists(ctx context.Context, channel string, msgID int) (bool, error) {
	filter := bson.M{"channel": channel, "msg_id": msgID}
	count, err := a.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (a *MongoAdapter) Close(ctx context.Context) error {
	return a.client.Disconnect(ctx)
}
