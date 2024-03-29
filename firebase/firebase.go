package firebase

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/storage"
	"github.com/tianhai82/ivsensor/model"
)

var config = &firebase.Config{
	StorageBucket: "ivsensor.appspot.com",
}
var app *firebase.App
var AuthClient *auth.Client
var StorageClient *storage.Client
var FirestoreClient *firestore.Client
var Stocks []model.Stock
var StockSymbols []string

func Init() {
	var err error
	ctx := context.Background()
	app, err = firebase.NewApp(ctx, config)
	if err != nil {
		log.Printf("firebase.NewApp failed: %v", err)
		return
	}
	AuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Printf("app.Auth failed: %v", err)
		return
	}
	StorageClient, err = app.Storage(ctx)
	if err != nil {
		log.Printf("app.Storage failed: %v", err)
		return
	}

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Printf("app.Firestore failed: %v", err)
		return
	}

	bucket, err := StorageClient.DefaultBucket()
	if err != nil {
		log.Printf("StorageClient.DefaultBucket() failed: %v", err)
		return
	}
	reader, err := bucket.Object("high_volume.json").NewReader(context.Background())
	if err != nil {
		log.Printf("bucket.Object(high_volume.json).NewReader failed: %v", err)
	} else {
		dec := json.NewDecoder(reader)
		var temp []model.Stock
		err = dec.Decode(&temp)
		if err != nil {
			println(err)
			return
		}
		for _, s := range temp {
			if s.AvgVolume90Day > 200 {
				Stocks = append(Stocks, s)
			}
		}
		println("number of stocks", len(Stocks))
	}

	reader2, err := bucket.Object("optionsSymbols.json").NewReader(context.Background())
	if err != nil {
		log.Printf("bucket.Object(optionsSymbols.json).NewReader failed: %v", err)
		return
	}
	dec2 := json.NewDecoder(reader2)

	err = dec2.Decode(&StockSymbols)
	if err != nil {
		println(err)
		return
	}
	println("number of stock symbols", len(StockSymbols))
}

func init() {
	Init()
}
