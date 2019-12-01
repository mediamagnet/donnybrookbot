package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jonas747/dca"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var err error
var slogan = "Donnybrook - Because sometimes fast needs to be quantified."
var race string
var startTime1 = time.Now()
var endTime1 = time.Now()
var wg sync.WaitGroup

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
	PlayersEntered int       `bson:"Players Entered"`
	PlayersReady   int       `bson:"Players Ready"`
	PlayersDone    int       `bson:"Players Done"`
}

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
func monPlayer(dbase string, collect string, players Players) {
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

func monRace(dbase string, collect string, races Races) {
	// Connecting to mongoDB
	client := GetClient()
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	}
	collection := client.Database(dbase).Collection(collect)
	_, _ = collection.InsertOne(context.TODO(), races)
}

func monReturnAllPlayers(client *mongo.Client, filter bson.M) []*Players {

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

func monReturnAllRaces(client *mongo.Client, filter bson.M) []*Races {

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

func monUpdatePlayer(client *mongo.Client, updatedData bson.M, filter bson.M) int64 {
	collection := client.Database("donnybrook").Collection("players")
	update := bson.D{{Key: "$set", Value: updatedData}}
	updatedResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("Error updating player", err)
	}
	return updatedResult.ModifiedCount
}

func monUpdateRace(client *mongo.Client, updatedData bson.M, filter bson.M) int64 {
	collection := client.Database("donnybrook").Collection("races")
	update := bson.D{{Key: "$set", Value: updatedData}}
	updatedResult, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("Error updating race,", err)
	}
	return updatedResult.ModifiedCount
}

func voiceChannels(s *discordgo.Session, guildID string) []string {
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

func findUserVoiceState(session *discordgo.Session, userID string) *discordgo.VoiceState {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userID {
				return vs
			}
		}
	}
	return nil
}

func findAllVoiceState(session *discordgo.Session) []string {
	vString := make([]string, 1)
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			vString = append(vString, vs.UserID)
		}
	}
	return vString
}

func currentVoiceChannel(session *discordgo.Session, userID string) []string {
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

func joinUserVoiceChannel(session *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
	// Find a user's current voice channel
	vs := findUserVoiceState(session, userID)
	if err != nil {
		log.Fatal("Error")
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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Test to make sure bot isn't talking to self.
	if m.Author.ID == s.State.User.ID {
		return
	}
	randomString(4)
	// Set up Race
	switch {

	case strings.HasPrefix(m.Content, ".setup"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		c := GetClient()
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
			lookupRace := monReturnAllRaces(c, bson.M{"RaceID": raceid})
			fmt.Println(raceid)
			for _, v := range lookupRace {
				if v.RaceID == raceid {
					raceid = randomString(4)
					race = raceid
				}
			}
			fmt.Println(raceid)
			racestring := fmt.Sprintf("%s, has started race %s for %s in category %s.", "<"+"@"+m.Author.ID+">", raceid, strings.TrimSpace(arg1), strings.TrimSpace(arg2))
			_, _ = s.ChannelMessageSend(m.ChannelID, racestring)
			raceInsert := Races{race, m.GuildID, m.ChannelID, strings.TrimSpace(arg1), strings.TrimSpace(arg2), time.Now(),0, 0, 0}
			fmt.Println(raceInsert)
			monRace("donnybrook", "races", raceInsert)

		}
	// Join Race
	case strings.HasPrefix(m.Content, ".join"):
		playerFound := ""
		raceFound := ""
		playersJoined := 0
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		c := GetClient()
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".join")
		raceID := strings.TrimSpace(msgstr)
		playerLookup := monReturnAllPlayers(c, bson.M{"RaceID": raceID})
		raceLookup := monReturnAllRaces(c, bson.M{"RaceID": raceID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerFound = v.PlayerID
			}
		}

		for _, v := range raceLookup {
			if v.RaceID == raceID {
				raceFound = v.RaceID
				playersJoined = v.PlayersEntered
			}
		}

		if playerFound == m.Author.ID {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already joined the race please wait for the race to start.")
		} else if len(raceID) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+">, I need your race ID also.")
		} else if raceID != race {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry that race id does not exist")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">"+" has joined the race.")
			playersJoined = playersJoined + 1
			logstring := fmt.Sprintf("Channel ID: %s, Race ID: %s, Name: %s", m.ChannelID, race, m.Author.ID)
			fmt.Println(logstring)
			player := Players{m.Author.Username, m.GuildID, m.Author.ID, m.ChannelID, raceID, false, false, time.Now(), time.Now()}
			monPlayer("donnybrook", "players", player)
			monUpdateRace(GetClient(), bson.M{"Players Entered": playersJoined}, bson.M{"RaceID": raceFound})
		}

	case strings.HasPrefix(m.Content, "a.scatter"):
		var voice *discordgo.VoiceConnection
		voice, _ = joinUserVoiceChannel(s, m.Author.ID)
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var msgstr = m.Content
		var msgarr = make([]string, 1)
		msgstr = strings.TrimPrefix(msgstr, "a.scatter ")
		msgarr = strings.Split(msgstr, ",")
		fmt.Println(msgarr)
		rand.Seed(time.Now().Unix())
		vUsers := findAllVoiceState(s)
		wg.Add(len(vUsers))
		// vChan := voiceChannels(s, m.GuildID)
		go PlayAudioFile(voice, "media/scatter2.mp3")
		for i := 0; i <= len(vUsers)-1; i++ {
			go func(i int) {
				defer wg.Done()
				vChan := rand.Intn(len(msgarr))
				pickedChan := ChannelIDFromName(s, m.GuildID, msgarr[vChan])
				choiceUser := vUsers[i]
				fmt.Println(pickedChan)
				vChan = rand.Intn(len(msgarr))
				pickedChan = ChannelIDFromName(s, m.GuildID, msgarr[vChan])

				_ = s.GuildMemberMove(m.GuildID, choiceUser, pickedChan)
				time.Sleep(500 * time.Millisecond)
			}(i)
		}
		wg.Wait()
	}

	switch {
	// Ready up for race
	case m.Content == ".ready":
		var playerReady bool
		playerRace := ""
		readyPlayer := 0
		c := GetClient()
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		playerLookup := monReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerReady = v.Ready
				playerRace = v.RaceID
			}
		}
		raceLookup := monReturnAllRaces(c, bson.M{"RaceID": playerRace})
		for _, v := range raceLookup {
			if v.RaceID == playerRace {
				readyPlayer = v.PlayersReady
			}
		}
		readyPlayer = readyPlayer + 1
		if playerReady == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already readied up for the race if you need to leave use `.unready`.")
		} else {
			monUpdatePlayer(c, bson.M{"Ready": true}, bson.M{"PlayerID": m.Author.ID})
			monUpdateRace(c, bson.M{"Players Ready":readyPlayer}, bson.M{"RaceID":playerRace})
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> is ready.")
		}
	// unready
	case m.Content == ".unready":
		var playerReady bool
		c := GetClient()
		playerLookup := monReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerReady = v.Ready
			}
		}
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if playerReady == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> has left ready status, please ready up again when able.")
			monUpdatePlayer(c, bson.M{"Ready": false}, bson.M{"PlayerID": m.Author.ID})
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> you have not readied up at all please .ready first.")
		}
	// Start Race once all ready
	case m.Content == ".start":

		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := joinUserVoiceChannel(s, m.Author.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "All racers have readied")
		time.Sleep(1 * time.Second)
		go s.ChannelMessageSend(m.ChannelID, "Starting in \n3.")
		go PlayAudioFile(voice, "media/racestart.mp3")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "2.")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "1.")
		time.Sleep(1 * time.Second)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Go!")

		startTime1 = time.Now()
		var starttime = startTime1.Truncate(1 * time.Millisecond)
		racestring := fmt.Sprintf("%s, %s, %s", m.ChannelID, race, starttime)
		fmt.Println(racestring)

	// You finish the Race
	case m.Content == ".done":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var timeJoined time.Time
		var playerDone bool
		c := GetClient()
		playerLookup := monReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				timeJoined = v.JoinTime
				playerDone = v.Done
			}
		}
		var time2 = time.Now()
		var endtime1 = time2.Sub(timeJoined)
		var endtime = endtime1.Truncate(1 * time.Millisecond)
		if playerDone == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You've already finished the race <@"+m.Author.ID+"> Please wait for the next race.")
		} else {
			monUpdatePlayer(c, bson.M{"Done Time": time2, "Done": true, "RaceID": "Done"}, bson.M{"PlayerID": m.Author.ID})
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
		}
	// Quit the race
	case m.Content == ".forfeit":
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has forfeit the race, hope you join us for the next race!")
	// Help text
	case m.Content == ".help":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "Donnybrook Help:",
			Author: &discordgo.MessageEmbedAuthor{
				URL:     "https://donnybrookbot.xyz",
				Name:    "Donnybrook",
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128"},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128",
				Width:  128,
				Height: 128,
			},
			Color:       0x550000,
			Description: "Welcome to the Donnybrook Race bot Here's some useful commands: \n",
			Fields: []*discordgo.MessageEmbedField{
				{Name: ".setup Game, Category", Value: "Setup a race"},
				{Name: ".join <race id>", Value: "Join a race with the specified id"},
				{Name: ".ready", Value: "Ready up for the race after you're done setting up"},
				{Name: ".unready", Value: "Leave the ready state for the race."},
				{Name: ".start", Value: "Once everyone is ready start the race."},
				{Name: ".done", Value: "Once you finish the race use this."},
				{Name: ".forfeit", Value: "Not able to keep going use this to exit the race."},
				{Name: ".help", Value: "You're reading it."},
				{Name: "v.join", Value: "Join a voice channel for voiced countdown. Must be in voice channel for this to work."},
				{Name: "v.leave", Value: "Leave the voice channel"},
				{Name: "a.cleanup", Value: "Cleans messages in channel command is run in. User must have `Mannage Message` permissions."}},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    slogan,
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=16"}})
	// Helf text
	case m.Content == ".helf":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "Donnybrook Helf:",
			Author: &discordgo.MessageEmbedAuthor{
				URL:     "https://donnybrookbot.xyz",
				Name:    "Donnybrook",
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128"},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128",
				Width:  128,
				Height: 128,
			},
			Color:       0x550000,
			Description: "Hmm, never mind.",
			Image: &discordgo.MessageEmbedImage{
				URL:      "https://i.imgur.com/d86T9Mw.png",
				ProxyURL: "https://i.imgur.com/d86T9Mw.png",
				Width:    428,
				Height:   559,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    slogan,
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=16"}})
	// Join voice
	case m.Content == "v.join":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = joinUserVoiceChannel(s, m.Author.ID)
	// Leave voice
	case m.Content == "v.shutup":
		// _ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

		// Clean up channel currently in.
	case m.Content == "a.cleanup":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

		if MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Deleting messages in <#"+m.ChannelID+">")
			time.Sleep(3 * time.Second)
			wg.Add(100)
			for i := 0; i < 100; i++ {

				go func(i int) {
					defer wg.Done()
					messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
					_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
				}(i)
			}
			wg.Wait()
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+"> You need Manage Message permissions to run a.cleanup")
		}
	// Bees
	case m.Content == "b.swarm":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "\n :bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n ")
	case m.Content == "g.honk":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := joinUserVoiceChannel(s, m.Author.ID)
		go s.ChannelMessageSend(m.ChannelID, "Peace was never an option \n https://i.kym-cdn.com/photos/images/newsfeed/001/597/651/360.jpg")
		go PlayAudioFile(voice, "media/honk.mp3")
		time.Sleep(10 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	}
}