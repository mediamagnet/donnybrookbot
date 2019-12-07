package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"math/rand"
	"time"
)
// TODO: Settings collection

// Players should have comment
type Players struct {
	Name      string    `bson:"Name"`
	GuildID	  string    `bson:"GuildID"`
	PlayerID  string    `bson:"PlayerID"`
	ChannelID string    `bson:"ChannelID"`
	RaceID    string    `bson:"RaceID"`
	Done	  bool      `bson:"Done"`
	Ready	  bool		`bson:"Ready"`
	JoinTime  time.Time `bson:"Join Time,omitempty"`
	DoneTime  time.Time `bson:"Done Time,omitempty"`
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

var err error

// GetClient
func GetClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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

// mongoDB connection stuff
func MonPlayer(dbase string, collect string, players Players) {
	// Connecting to mongoDB
	client := GetClient()
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	}
	collection := client.Database(dbase).Collection(collect)

	insertResult, err := collection.InsertOne(context.TODO(), players)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted:", insertResult.InsertedID)
}

func MonRace(dbase string, collect string, races Races) {
	// Connecting to mongoDB
	client := GetClient()
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	}
	collection := client.Database(dbase).Collection(collect)
	insertResult, err := collection.InsertOne(context.TODO(), races)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted:", insertResult.InsertedID)
}

func MonReturnAllPlayers(client *mongo.Client, filter bson.M) []*Players {

	var players []*Players
	collection := client.Database("donnybrook").Collection("players")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal("Could not find document ", err)
	}
	for cur.Next(context.TODO()) {
		var player Players
		err = cur.Decode(&player)
		if err != nil {
			log.Fatal("Decode Error ", err)
		}
		players = append(players, &player)
	}
	return players
}

func MonReturnAllRaces(client *mongo.Client, filter bson.M) []*Races {

	var races []*Races
	collection := client.Database("donnybrook").Collection("races")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal("Could not find document ", err)
	}
	for cur.Next(context.TODO()) {
		var race Races
		err = cur.Decode(&race)
		if err != nil {
			log.Fatal("Decode Error ", err)
		}
		races = append(races, &race)
	}
	return races
}

func MonUpdatePlayer(client *mongo.Client, updatedData bson.M, filter bson.M) int64 {
	collection := client.Database("donnybrook").Collection("players")
	update := bson.D{{Key: "$set", Value: updatedData}}
	updatedResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("Error updating player", err)
	}
	return updatedResult.ModifiedCount
}

func MonUpdateRace(client *mongo.Client, updatedData bson.M, filter bson.M) int64 {
	collection := client.Database("donnybrook").Collection("races")
	update := bson.D{{Key: "$set", Value: updatedData}}
	updatedResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("Error updating race,", err)
	}
	return updatedResult.ModifiedCount
}

func VoiceChannels(s *discordgo.Session, guildID string) []string {
	channels, _ := s.GuildChannels(guildID)
	chanSlice := make([]string, 1)
	for _, c := range channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			chanID := fmt.Sprintf("%s", c.ID)
			chanSlice = append(chanSlice, chanID)
		}
	}
	return chanSlice
}

func ChannelIDFromName(s *discordgo.Session, guildID string, channelName string) string {
	channels, _ := s.GuildChannels(guildID)
	var c = ""
	for _, cList := range channels {
		if cList.Name == channelName {
			c = cList.ID
		}
	}
	return c
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(RandomInt(65, 90))
	}
	return string(bytes)
}

func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int) bool {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false
		}
	}

	// Iterate through the role IDs stored in member.Roles
	// to check permissions
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}

	return false
}

func FindUserVoiceState(session *discordgo.Session, userID string) (*discordgo.VoiceState, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userID {
				return vs, nil
			}
		}
	}
	return nil, errors.New("Could not find user's voice state")
}

func FindAllVoiceState(session *discordgo.Session) []string {
	vString := make([]string, 1)
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			vString = append(vString, vs.UserID)
		}
	}
	return vString
}

func CurrentVoiceChannel(session *discordgo.Session, userID string) []string {
	var vUser = make([]string, 1)
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userID {
				vUser = append(vUser, vs.ChannelID)
			}
		}
	}
	return vUser
}

func JoinUserVoiceChannel(session *discordgo.Session, channelID string, userID string) (*discordgo.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := FindUserVoiceState(session, userID)
	if err != nil {
		session.ChannelMessageSend(channelID,"Error")
	}
	//
	return session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, false)
}

func PlayAudioFile(v *discordgo.VoiceConnection, filename string) {

	// Send "speaking" packet over the voice websocket
	err := v.Speaking(true)
	if err != nil {
		log.Fatal("Failed setting speaking", err)
	}

	// Send not "speaking" packet over the websocket when we finish
	defer v.Speaking(false)

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	encodeSession, err := dca.EncodeFile(filename, opts)
	if err != nil {
		log.Fatal("Failed creating an encoding session: ", err)
	}

	done := make(chan error)
	stream := dca.NewStream(encodeSession, v, done)

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Fatal("An error occured", err)
			}

			// Clean up incase something happened and ffmpeg is still running
			encodeSession.Truncate()
			return
		case <-ticker.C:
			stats := encodeSession.Stats()
			playbackPosition := stream.PlaybackPosition()

			fmt.Printf("Playback: %10s, Transcode Stats: Time: %5s, Size: %5dkB, Bitrate: %6.2fkB, Speed: %5.1fx\r", playbackPosition, stats.Duration.String(), stats.Size, stats.Bitrate, stats.Speed)
		}
	}
}