package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jonas747/dca"
	"go.mongodb.org/mongo-driver/bson"
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
	PlayerID  string    `bson:"PlayerID"`
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

func monReturnOnePlayer(client *mongo.Client, filter bson.M) Players {
	var player Players
	collection := client.Database("donnybrook").Collection("players")
	docuReturned := collection.FindOne(context.TODO(), filter)
	_ = docuReturned.Decode(&player)
	return player
}

func voiceChannels(s *discordgo.Session, guildID string) []string {
	channels, _ := s.GuildChannels(guildID)
	for _, c := range channels {
		if c.Type != discordgo.ChannelTypeGuildVoice {
			continue
		}
		chanString := fmt.Sprintf("%s", c.ID)
		chanSlice := strings.Fields(chanString)
		return chanSlice

	}
	return nil
}

func ChannelIDFromName(s *discordgo.Session, guildID string, channelName string) string {
	channels, _ := s.GuildChannels(guildID)

	for _, c := range channels {
		if c.Name == channelName {
			return c.ID
		}
	}
	return ""
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

func findUserVoiceState(session *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("could not find user's voice state")
}

func findAllVoiceState(session *discordgo.Session) (string, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			vString := fmt.Sprintf("%s", vs)
			return vString, nil
		}
	}
	return "", errors.New("could not find user's voice state")
}

func joinUserVoiceChannel(session *discordgo.Session, userID string) (*discordgo.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := findUserVoiceState(session, userID)
	if err != nil {
		return nil, err
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

	// Set up Race
	switch {

	case strings.HasPrefix(m.Content, ".setup"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
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
			monRace("donnybrook", "races", raceInsert)

		}
	// Join Race
	case strings.HasPrefix(m.Content, ".join"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		// c := GetClient()
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".join")
		raceID := strings.TrimSpace(msgstr)
		// playerLookup := monReturnOnePlayer(c, bson.M{"PlayerID": m.Author.ID})
		playerLookup := staticuser
		if playerLookup == m.Author.ID {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already joined the race please wait for the race to start.")
		} else if len(raceID) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+">, I need your race ID also.")
		} else if raceID != race {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry that race id does not exist")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">"+" has joined the race.")
			logstring := fmt.Sprintf("Channel ID: %s, Race ID: %s, Name: %s", m.ChannelID, race, m.Author.Username)
			fmt.Println(logstring)
			player := Players{m.Author.Username, m.Author.ID, m.ChannelID, raceID, time.Now(), time.Now()}
			monPlayer("donnybrook", "players", player)
		}
	case strings.HasPrefix(m.Content, ".inrace"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		c := GetClient()
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".inrace")
		raceID := strings.TrimSpace(msgstr)
		if len(raceID) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+">, I need the race ID also.")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Players in the race are: ")
			enteredPlayers := monReturnAllPlayers(c, bson.M{"RaceID": raceID})
			for _, player := range enteredPlayers {
				msgString := fmt.Sprintf("Name: %s, Channel: %s, Race: %s", player.Name, player.ChannelID, player.RaceID)
				_, _ = s.ChannelMessageSend(m.ChannelID, msgString)
			}

		}
		//case strings.HasPrefix(m.Content, "a.scatter"):
		//	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		//	var msgstr = m.Content
		//	msgstr = strings.TrimPrefix(msgstr, "a.scatter")
		//	channelName := strings.TrimSpace(msgstr)
		//	rand.Seed(time.Now().UnixNano())
		//	s.Guild
		//
		//	vChans := strings.Fields(voiceChannels(s))
		//	choosen := vChans[rand.Intn(len(vChans))]
		//	s.GuildMemberMove(m.GuildID,,choosen)
		//
	}
	switch {
	// Ready up for race
	case m.Content == ".ready":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if m.Author.ID == staticuser {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" You've already readied up for the race if you need to leave use `.unready`.")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" is ready.")
		}
	// unready
	case m.Content == ".unready":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has left ready status, please ready up again when able.")
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

		starttime1 = time.Now()
		var starttime = starttime1.Truncate(1 * time.Millisecond)
		racestring := fmt.Sprintf("%s, %s, %s", m.ChannelID, race, starttime)
		fmt.Println(racestring)

	// You finish the Race
	case m.Content == ".done":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var time2 = time.Now()
		var endtime1 = time2.Sub(starttime1)
		var endtime = endtime1.Truncate(1 * time.Millisecond)
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
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
				{Name: "v.join", Value: "Join a voice channel for voiced countdown. Must be in voice channel for this to work. [WIP]"},
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
	case m.Content == "v.leave":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)
	// Clean up channel currently in.
	case m.Content == "a.cleanup":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Deleting messages in <#"+m.ChannelID+">")
			msgCount, _ := s.ChannelMessages(m.ChannelID, 100, "", "", "")
			time.Sleep(5 * time.Second)
			if len(msgCount) == 100 {
				for i := 0; i < 100; i++ {
					messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
					_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
					time.Sleep(5 * time.Millisecond)
				}
			} else if len(msgCount) <= 99 {
				for i := 0; i < len(msgCount); i++ {
					messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
					_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
					time.Sleep(5 * time.Millisecond)
				}
			}
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+"> You need Manage Message permissions to run a.cleanup")
		}
	// Bees
	case m.Content == "b.swarm":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "\n :bee: :bee: :bee: :bee: \n :bee: :bee: :bee: :bee: \n :bee: :bee: :bee: :bee: \n :bee: :bee: :bee: :bee: \n ")
	}
}



