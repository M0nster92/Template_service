package main

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBConnect(URI string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(URI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Problem connecting to Mongo URI %v, received errer %v", URI, err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Problem pinging server %v, received error %v", URI, err)
	} else {
		log.Infof("Connected to MongoDB %v", URI)
	}
	return client
}

func GetNextCode(collection *mongo.Collection, field string, prefix string) string {
	prefixLength := len(prefix)
	findOptions := options.Find()
	findOptions.SetLimit(1)
	findOptions.SetCollation(&options.Collation{Locale: "en_US", NumericOrdering: true})
	findOptions.SetSort(bson.D{{field, -1}})
	findOptions.SetProjection(bson.M{field: 1, "codefield": "$" + field})

	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	var ccint int

	if cur.Next(context.TODO()) {
		var record map[string]interface{}

		err = cur.Decode(&record)
		if err != nil {
			log.Error(err)
		}
		if record[field] != nil {
			code := record[field].(string)
			ccint, _ = strconv.Atoi(code[prefixLength:])
			log.Infof("Last record number is %d", ccint)
			ccint += 1
		} else {
			ccint = 1
			log.Infof("no records for %v - record is nil, starting at 1", field)
		}
	} else {
		ccint = 1
		log.Infof("no records for %v - starting at 1", field)
	}
	cur.Close(context.TODO())
	nextcode := fmt.Sprintf(prefix+"%05d", ccint)
	log.Infof("Next code will be %v", nextcode)
	return nextcode

}

type Filter struct {
	Key   string
	Value string
}
