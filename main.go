package main

import (
	"donnybrook/tools"
	"fmt"
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
	"go.mongodb.org/mongo-driver/bson"
)

var slogan = "Donnybrook - Because sometimes fast needs to be quantified."
var race string
var startTime1 = time.Now()
// var endTime1 = time.Now()
var wg sync.WaitGroup

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


func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Test to make sure bot isn't talking to self.
	if m.Author.ID == s.State.User.ID {
		return
	}
	tools.RandomString(4)
	// Set up Race
	switch {

	case strings.HasPrefix(m.Content, ".setup"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		c := tools.GetClient()
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".setup")
		msgstr = strings.TrimSpace(msgstr)
		msgarr := strings.Split(msgstr, ",")

		if len(msgarr) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You're missing either the game or the category of the race. run `.setup game name, category`")
		} else {
			var arg1 = msgarr[0]
			var arg2 = msgarr[1]
			var raceid = tools.RandomString(4)
			race = raceid
			lookupRace := tools.MonReturnAllRaces(c, bson.M{"RaceID": raceid})
			fmt.Println(raceid)
			for _, v := range lookupRace {
				if v.RaceID == raceid {
					raceid = tools.RandomString(4)
					race = raceid
				}
			}
			fmt.Println(raceid)
			racestring := fmt.Sprintf("%s, has started race %s for %s in category %s.", "<"+"@"+m.Author.ID+">", raceid, strings.TrimSpace(arg1), strings.TrimSpace(arg2))
			_, _ = s.ChannelMessageSend(m.ChannelID, racestring)
			raceInsert := tools.Races{RaceID: race,
									  GuildID: m.GuildID,
									  ChannelID: m.ChannelID,
									  Game: strings.TrimSpace(arg1),
									  Category: strings.TrimSpace(arg2),
									  StartTime: time.Now()}
			fmt.Println(raceInsert)
			tools.MonRace("donnybrook", "races", raceInsert)

		}
	// Join Race
	case strings.HasPrefix(m.Content, ".join"):
		playerFound := ""
		raceFound := ""
		playersJoined := 0
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		c := tools.GetClient()
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".join")
		raceID := strings.TrimSpace(msgstr)
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"RaceID": raceID})
		raceLookup := tools.MonReturnAllRaces(c, bson.M{"RaceID": raceID})
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
			player := tools.Players{Name: m.Author.Username,
									GuildID: m.GuildID,
									PlayerID: m.Author.ID,
									ChannelID: m.ChannelID,
									RaceID: raceID,
									Ready: false,
									Done: false,
									JoinTime: time.Now(),
									DoneTime: time.Now()}

			tools.MonPlayer("donnybrook", "players", player)
			tools.MonUpdateRace(tools.GetClient(), bson.M{"Players Entered": playersJoined}, bson.M{"RaceID": raceFound})
		}

	case strings.HasPrefix(m.Content, "a.scatter"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var voice *discordgo.VoiceConnection
		voice, _ = tools.JoinUserVoiceChannel(s, m.Author.ID)
		var msgstr = m.Content
		var msgarr = make([]string, 1)
		msgstr = strings.TrimPrefix(msgstr, "a.scatter ")
		msgarr = strings.Split(msgstr, ",")
		fmt.Println(msgarr)
		rand.Seed(time.Now().Unix())
		vUsers := tools.FindAllVoiceState(s)
		if tools.MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageServer|discordgo.PermissionAll) {
			wg.Add(len(vUsers))
			// vChan := voiceChannels(s, m.GuildID)
			go tools.PlayAudioFile(voice, "media/scatter2.mp3")
			for i := 0; i <= len(vUsers)-1; i++ {
				go func(i int) {
					defer wg.Done()
					vChan := rand.Intn(len(msgarr))
					pickedChan := tools.ChannelIDFromName(s, m.GuildID, msgarr[vChan])
					choiceUser := vUsers[i]
					fmt.Println(pickedChan)
					vChan = rand.Intn(len(msgarr))
					pickedChan = tools.ChannelIDFromName(s, m.GuildID, msgarr[vChan])

					_ = s.GuildMemberMove(m.GuildID, choiceUser, pickedChan)
					time.Sleep(500 * time.Millisecond)
				}(i)
			}
			wg.Wait()
			} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "No <@"+m.Author.ID+"> you cannot scatter people.")
		}
	}

	switch {
	// Ready up for race
	case m.Content == ".ready":
		var playerReady bool
		playerRace := ""
		readyPlayer := 0
		c := tools.GetClient()
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerReady = v.Ready
				playerRace = v.RaceID
			}
		}
		raceLookup := tools.MonReturnAllRaces(c, bson.M{"RaceID": playerRace})
		for _, v := range raceLookup {
			if v.RaceID == playerRace {
				readyPlayer = v.PlayersReady
			}
		}
		readyPlayer = readyPlayer + 1
		if playerReady == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already readied up for " +
				"the race if you need to leave use `.unready`.")
		} else {
			tools.MonUpdatePlayer(c, bson.M{"Ready": true}, bson.M{"PlayerID": m.Author.ID})
			tools.MonUpdateRace(c, bson.M{"Players Ready":readyPlayer}, bson.M{"RaceID":playerRace})
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> is ready.")
		}
	// unready
	case m.Content == ".unready":
		var playerReady bool
		c := tools.GetClient()
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerReady = v.Ready
			}
		}
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if playerReady == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> has left ready status, please ready " +
				"up again when able.")
			tools.MonUpdatePlayer(c, bson.M{"Ready": false}, bson.M{"PlayerID": m.Author.ID})
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> you have not readied up at all " +
				"please .ready first.")
		}
	// Start Race once all ready
	case m.Content == ".start":

		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.Author.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "All racers have readied")
		time.Sleep(1 * time.Second)
		go s.ChannelMessageSend(m.ChannelID, "Starting in \n3.")
		go tools.PlayAudioFile(voice, "media/racestart.mp3")
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
		c := tools.GetClient()
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
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
			_, _ = s.ChannelMessageSend(m.ChannelID, "You've already finished the race <@"+m.Author.ID+"> " +
				"Please wait for the next race.")
		} else {
			tools.MonUpdatePlayer(c, bson.M{"Done Time": time2, "Done": true, "RaceID": "Done"}, bson.M{"PlayerID": m.Author.ID})
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
		}
	// Quit the race
	case m.Content == ".forfeit":
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has forfeit the race, " +
			"hope you join us for the next race!")
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
				{Name: "a.cleanup", Value: "Cleans messages in channel command is run in. User must have `Mannage Message` permissions."},
				{Name: "a.scatter <Channels to scatter to>", Value:"Scatter users to the provided channels. User must have `Manage Server` Permissions."},
				{Name: ".lick", Value: "..."}},
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
		_, _ = tools.JoinUserVoiceChannel(s, m.Author.ID)
	// Leave voice
	case m.Content == "v.leave":
		// _ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

		// Clean up channel currently in.
	case m.Content == "a.cleanup":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

		if tools.MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) {
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
		lastMsg :=  make([]string, 1)
		fmt.Println(lastMsg)
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "\n :bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n ")
	case m.Content == "g.honk":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.Author.ID)
		go s.ChannelMessageSend(m.ChannelID, "Peace was never an option \n " +
			"https://i.kym-cdn.com/photos/images/newsfeed/001/597/651/360.jpg")
		go tools.PlayAudioFile(voice, "media/honk.mp3")
		time.Sleep(10 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	case m.Content == ".lick":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.Author.ID)
		go s.ChannelMessageSend(m.ChannelID, "https://i.ytimg.com/vi/lSXxEdaOqgU/maxresdefault.jpg")
		go tools.PlayAudioFile(voice, "media/lick.mp3")
		time.Sleep(50 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	}
}