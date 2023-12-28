package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	model "github.com/TealWater/clear-scribe/src/Model"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var database *mongo.Database
var mongoClient *mongo.Client

const dbName = "notes"
const colName = "messages"

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("unable to laod environment variables")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoURI := "mongodb+srv://" + os.Getenv("MONGO_DB_USERNAME") + ":" + os.Getenv("MONGO_DB_PASSWORD") + "@cluster0.lx82yxi.mongodb.net/?retryWrites=true&w=majority"
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	mongoClient, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Println("unable to connect to DB")
		log.Panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := mongoClient.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")

	databases, err := mongoClient.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal("line 57!: ", err)
	}
	log.Println(databases)

	database = mongoClient.Database(dbName)
	collection = database.Collection(colName)
	fmt.Println("collection name: ", collection.Name())
}

func insertMessages(messageOld, messageNew string) {
	const format = "Jan 2, 2006 at 3:04pm (MST)"
	date := time.Now().Local()

	entry := model.EditedEssay{
		CreatedAt:  primitive.NewDateTimeFromTime(date),
		DateString: date.Format(format),
		MessageOld: messageOld,
		MessageNew: messageNew,
	}

	inserted, err := collection.InsertOne(context.TODO(), entry)
	if err != nil {
		log.Println("unable to insert data")
		log.Fatal(err)
	}

	log.Println("Insereted new row entry with the id of: ", inserted.InsertedID)
}

func deleteMessage(messageId string) error {
	id, _ := primitive.ObjectIDFromHex(messageId)
	filter := bson.M{"_id": id}
	deleteCount, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Println("unable to delete record from the database")
		return err
	}

	if deleteCount.DeletedCount == 0 {
		return errors.New("message with id " + messageId + " does not exist in the database")
	}

	log.Println("message with id ", messageId, " was deleted, with count of: ", deleteCount.DeletedCount)
	return nil
}

func deleteAllMessages() error {
	deleteCount, err := collection.DeleteMany(context.Background(), bson.D{{}}, nil)
	if err != nil {
		log.Println("Unable to delete all records")
		return err
	}

	log.Println("we deleted ", deleteCount.DeletedCount, " records")
	return nil
}

func getAllMessages() []primitive.M {
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Println("Unable to retreive all of the records")
		log.Fatalln(err)
	}

	var messages []primitive.M
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var message bson.M
		if err := cur.Decode(&message); err != nil {
			log.Println("unable to decode old message")
			log.Fatalln(err)
		}
		messages = append(messages, message)
	}
	return messages
}

func CloseDB() {
	if err := mongoClient.Disconnect(context.TODO()); err != nil {
		log.Println("error trying to disconnect from the DB")
		log.Panic(err)
	}
}
