package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var staticuser = "265562769397514242"
var err error
var atime = time.Now()
var slogan = "Donnybrook - Because sometimes fast needs to be quantified."
var race string
var starttime1 = time.Now()
var endtime1 = time.Now()

// Players should have comment
type Players struct {
	Name      string    `bson:"Name"`
	ChannelID string    `bson:"ChannelID"`
	RaceID    string    `bson:"RaceID"`
	JoinTime  time.Time `bson:"Join Time,omitempty"`
	DoneTime  time.Time `bson:"Done Time,omitempty"`
}

// Races should have comment
type Races struct {
	RaceID    string    `bson:"RaceID"`
	ChannelID string    `bson:"ChannelID"`
	Game      string    `bson:"Game"`
	Category  string    `bson:"Category"`
	StartTime time.Time `bson:"Start Time"`
}

// mongoDB connection stuff
func monplayer(dbase string, collect string, players Players) {
	// Connecting to mongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	fmt.Println("clientOptions type:", reflect.TypeOf(clientOptions))

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(dbase).Collection(collect)

	insertResult, err := collection.InsertOne(context.TODO(), players)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted:", insertResult.InsertedID)
}

func monrace(dbase string, collect string, races Races) {
	// Connecting to mongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	fmt.Println("clientOptions type:", reflect.TypeOf(clientOptions))

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(dbase).Collection(collect)
	_, _ = collection.InsertOne(context.TODO(), races)
}

func monLookupPlayer(dbase string, collect string, RaceID string) {

	filter := bson.D{{"RaceID", RaceID}}

	var result Players
	// Connecting to mongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	fmt.Println("clientOptions type:", reflect.TypeOf(clientOptions))

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbase).Collection(collect)
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil{
		log.Fatal(err)
	}
	fmt.Printf("Found it: \n", result)
}

func main() {
	// Load .env files
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to discord using token from ENV file
	dstoken := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + dstoken)
	if err != nil {
		log.Fatal("Error creating Discord session,", err)
		return
	}
	// Let's message things
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = dg.Close()

}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Test to make sure bot isn't talking to self.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Set up Race
	switch {

	case strings.HasPrefix(m.Content, ".setup"):
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".setup")
		msgstr = strings.TrimSpace(msgstr)
		msgarr := strings.Split(msgstr, ",")
		if len(msgarr) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You're missing either the game or the category of the race. run `.setup game name, category`")
		} else {
			var arg1 = msgarr[0]
			var arg2 = msgarr[1]
			var raceid = randomString(4)
			race = raceid
			racestring := fmt.Sprintf("%s, has started race %s for %s in category %s.", "<"+"@"+m.Author.ID+">", raceid, strings.TrimSpace(arg1), strings.TrimSpace(arg2))
			_, _ = s.ChannelMessageSend(m.ChannelID, racestring)
			raceInsert := Races{race, m.ChannelID, strings.TrimSpace(arg1), strings.TrimSpace(arg2), time.Now()}
			fmt.Println(raceInsert)
			monrace("donnybrook", "races", raceInsert)
		}
	// Join Race
	case strings.HasPrefix(m.Content, ".join"):
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".join")
		raceID := strings.TrimSpace(msgstr)
		if m.Author.ID == staticuser {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already joined the race please wait for the race to start.")
		} else if len(raceID) <=1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+">, I need your race ID also.")
		} else if raceID != race {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry that race id does not exist")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">"+" has joined the race.")
			logstring := fmt.Sprintf("Channel ID: %s, Race ID: %s, Name: %s", m.ChannelID, race, m.Author.Username)
			fmt.Println(logstring)
			player := Players {m.Author.ID, m.ChannelID,raceID, time.Now(), time.Now()}
			monplayer("donnybrook", "players", player)
		}
	case strings.HasPrefix(m.Content, ".inrace"):
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".inrace")
		raceID := strings.TrimSpace(msgstr)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Check console, programmer lazy.")
		monLookupPlayer("donnybrook", "players", raceID)
	}
	switch {
	// Ready up for race
	case m.Content == ".ready":
		if m.Author.ID == staticuser {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" You've already readied up for the race if you need to leave use `.unready`.")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" is ready.")
		}
	// unready
	case m.Content == ".unready":
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has left ready status, please ready up again when able.")
	// Start Race once all ready
	case m.Content == ".start":
		_, _ = s.ChannelMessageSend(m.ChannelID, "All racers have readied")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Starting in 3.")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "2.")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "1.")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Go!")
		starttime1 = time.Now()
		var starttime = starttime1.Truncate(1 * time.Millisecond)
		racestring := fmt.Sprintf("%s, %s, %s", m.ChannelID, race, starttime)
		fmt.Println(racestring)
	// You finish the Race
	case m.Content == ".done":
		var time2 = time.Now()
		var endtime1 = time2.Sub(starttime1)
		var endtime = endtime1.Truncate(1 * time.Millisecond)
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
	// Quit the race
	case m.Content == ".forfeit":
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has forfeit the race, hope you join us for the next race!")
	// Help text
	case m.Content == ".help":
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "Donnybrook Help:",
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128",
				Width:  128,
				Height: 128,
			},
			Description: "Welcome to the Donnybrook Race bot to get started use: `.setup` to create a race. \n" +
				"Example: ```.setup A Link to the Past, Any%``` \n" +
				"`.join` - Join the race. \n" +
				"`.ready` - Ready up for the race. \n " +
				"`.unready` - Unready from the race. \n" +
				"`.start` - Start the race once everyone is ready. \n" +
				"`.done` - You finished the race. \n" +
				"`.forfeit` - Leave the race early.\n" +
				"`.help` - You're reading it.",
			Color: 0x550000,
			Author: &discordgo.MessageEmbedAuthor{
				URL:     "https://donnybrookbot.xyz",
				Name:    "Donnybrook",
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128"},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    slogan,
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=16"}})
	}

}
