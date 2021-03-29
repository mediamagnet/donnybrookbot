package tools

import (
	"context"
	"donnybrook/config"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// Role blah

type Records struct {
	Role Role
	Players Players
	Races Races
	Settings Settings
}

type Role struct {
	GuildID   string `bson:"GuildID,omitempty"`   // GuildID to watch in
	ChannelID string `bson:"ChannelID,omitempty"` // ChannelID for where to watch for Phrase
	RoleID    string `bson:"RoleID,omitempty"`    // RoleID to assign
	IgnoreID  string `bson:"IgnoreID,omitempty"`  // IgnoreID ignore this role when logging in/ logging out
	Phrase    string `bson:"Phrase,omitempty"`    // Phrase to watch for to assign role
}

type Players struct {
	Name      string    `bson:"Name"`
	GuildID   string    `bson:"GuildID"`
	PlayerID  string    `bson:"PlayerID"`
	ChannelID string    `bson:"ChannelID"`
	RaceID    string    `bson:"RaceID"`
	Done      bool      `bson:"Done"`
	Ready     bool      `bson:"Ready"`
	JoinTime  time.Time `bson:"Join Time,omitempty"`
	DoneTime  time.Time `bson:"Done Time,omitempty"`
	TotalTime time.Time `bson:"Total Time,omitempty"`
}

// Races should have comment
type Races struct {
	RaceID         string    `bson:"RaceID"`
	GuildID        string    `bson:"GuildID"`
	ChannelID      string    `bson:"ChannelID"`
	Game           string    `bson:"Game"`
	Category       string    `bson:"Category"`
	StartTime      time.Time `bson:"Start Time"`
	Started        bool      `bson:"Started"`
	PlayersEntered int       `bson:"Players Entered"`
	PlayersReady   int       `bson:"Players Ready"`
	PlayersDone    int       `bson:"Players Done"`
}

type Settings struct {
	GuildID string `bson:"GuildID"`
	Volume  int    `bson:"Volume"`
	VoidChan string `bson:"VoidChan,omitempty"`
}

// GetClient blah
func GetClient() *mongo.Client {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	var cfg config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Fatalf("unable to decode into struct, %v", err)
	}

	mongoURL :=
		fmt.Sprintf("mongodb://%v:27017",
			cfg.Mongo.DB_URL)
	fmt.Println(mongoURL)
	clientOptions := options.Client().ApplyURI(mongoURL)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// MonRecord mongoDB connection stuff
func MonRecord(dbase string, collect string, listens Role) {
	// Connecting to mongoDB
	client := GetClient()
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	}
	collection := client.Database(dbase).Collection(collect)
	insertResult, err := collection.InsertOne(context.TODO(), listens)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted:", insertResult.InsertedID)
}

func MonDeleteRecord(client *mongo.Client, filter bson.M, dbase string, collect string) int64 {
	collection := client.Database(dbase).Collection(collect)
	deleteResult, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal("Error deleting Record", err)
	}
	return deleteResult.DeletedCount
}

// MonUpdateRecord blah
func MonUpdateRecord(client *mongo.Client, updatedData bson.M, filter bson.M, dbase string, collect string) int64 {
	collection := client.Database(dbase).Collection(collect)
	update := bson.D{{Key: "$set", Value: updatedData}}
	updatedResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("Error updating player", err)
	}
	return updatedResult.ModifiedCount
}

// MonReturnOneRecord blah
func MonReturnOneRecord(client *mongo.Client, filter bson.M, dbase string, collect string) Records {
	var phrase Records
	collection := client.Database(dbase).Collection(collect)
	documentReturned := collection.FindOne(context.TODO(), filter)
	documentReturned.Decode(&phrase)
	return phrase
}

// MonReturnAllRecords blah
func MonReturnAllRecords(client *mongo.Client, filter bson.M, dbase string, collect string) []*Records {
	var roles []*Records
	collection := client.Database(dbase).Collection(collect)
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal("Error finding all the things", err)
	}
	for cur.Next(context.TODO()) {
		var role Records
		err = cur.Decode(&role)
		if err != nil {
			log.Fatal("Error Decoding :( ", err)
		}
		roles = append(roles, &role)
	}
	return roles
}

