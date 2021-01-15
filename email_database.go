package main

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTemplate(collection *mongo.Collection, id string, application string) Template {
	log.Infof("Fetching e-mail template %v", id)
	filter := bson.M{"$and": []bson.M{
		bson.M{"template_id": id},
		bson.M{"Application": application},
	}}
	var templateData Template

	//	options := options.Find()
	result := collection.FindOne(context.TODO(), filter)
	err := result.Decode(&templateData)
	if err != nil {
		log.Errorf("Error decoding BSON %v from template %v", err, id)
	}

	return templateData
}

func (template *Template) Insert(database *mongo.Database) (success bool, err error) {

	template.CreatedDate = time.Now()
	result, err := database.Collection("emailTemplates").InsertOne(context.TODO(), template)
	if result.InsertedID != "" {
		success = true
	}

	return
}

func (template *Template) Update(database *mongo.Database) (success bool, err error) {

	filter := bson.D{{
		"template_id", template.ID,
	}}

	template.LastUpdate = time.Now()
	pByte, err := bson.Marshal(template)
	if err != nil {
		return
	}

	var update bson.M
	err = bson.Unmarshal(pByte, &update)
	if err != nil {
		return
	}

	var result *mongo.UpdateResult
	result, err = database.Collection("emailTemplates").UpdateOne(context.TODO(), filter, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		return
	}
	if result.ModifiedCount == 1 {
		log.Info("Template Updated")
		success = true
	} else if result.UpsertedCount == 1 {
		log.Error("Updated template with no Existing entry!")
		err = errors.New("Updated template with no existing entry")
	}

	return
}
