package firebase

import (
	"context"
	"encoding/json"
	"fmt"

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

func init() {
	var err error
	ctx := context.Background()
	app, err = firebase.NewApp(ctx, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	AuthClient, err = app.Auth(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	StorageClient, err = app.Storage(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	bucket, err := StorageClient.DefaultBucket()
	if err != nil {
		fmt.Println(err)
		return
	}
	reader, err := bucket.Object("high_volume.json").NewReader(context.Background())
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
	println(len(Stocks))
}
