package racebot

import (
	"donnybrook/tools"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
	"time"
)

var startTime1 = time.Now()
// var endTime1 = time.Now()
var wg sync.WaitGroup

func RaceBot(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			race := raceid
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

			_, err := s.GuildChannelCreateComplex(m.GuildID, discordgo.GuildChannelCreateData{
				Name:                 raceid,
				Type:                 discordgo.ChannelTypeGuildText,
				Topic:                arg1+" - "+arg2,
			})
			_, err = s.GuildChannelCreateComplex(m.GuildID, discordgo.GuildChannelCreateData{
				Name:                 raceid+"-voice",
				Type:                 discordgo.ChannelTypeGuildVoice,
				Bitrate:              64000,
			})
			if err != nil {
				fmt.Println("Error creating channel", err)
			}
			getID := tools.ChannelIDFromName(s, m.GuildID, raceid+"-voice")
			_, _ = s.ChannelVoiceJoin(m.GuildID, getID, false, false)

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
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
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
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> You've already joined a race please wait for that race to start or .leave to join a different one.")
		} else if len(raceID) <= 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+">, I need your race ID also.")
		} else if raceID != raceFound {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry that race id does not exist, or you are in a different race already.")
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">"+" has joined the race.")
			playersJoined = playersJoined + 1
			logstring := fmt.Sprintf("Channel ID: %s, Race ID: %s, Name: %s", m.ChannelID, raceID, m.Author.ID)
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
			// getID := tools.ChannelIDFromName(s, m.GuildID, raceFound+"-voice")
			// _ = s.GuildMemberMove(m.GuildID, m.Author.ID, getID)
		}
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
		var raceID string
		var readyCount int
		c := tools.GetClient()
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				playerReady = v.Ready
				raceID = v.RaceID
			}
		}
		raceLookup := tools.MonReturnAllRaces(c, bson.M{"RaceID": raceID})
		for _, v := range raceLookup {
			if v.RaceID == raceID {
				readyCount = v.PlayersReady
			}
		}
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if playerReady == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> has left ready status, please ready " +
				"up again when able.")
			tools.MonUpdatePlayer(c, bson.M{"Ready": false}, bson.M{"PlayerID": m.Author.ID})
			tools.MonUpdateRace(c, bson.M{"Players Ready": readyCount-1}, bson.M{"RaceID": raceID})
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+"> you have not readied up at all " +
				"please .ready first.")
		}
	// Start Race once all ready
	case m.Content == ".start":

		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		startTime1 = time.Now()
		var starttime = startTime1.Truncate(1 * time.Millisecond)
		var getRace string
		var readyCount int
		var joinCount int
		allPlayers := tools.MonReturnAllPlayers(tools.GetClient(), bson.M{"PlayerID": m.Author.ID})
		for _, v := range allPlayers {
			if v.PlayerID == m.Author.ID {
				if v.RaceID != "done" {
					getRace = v.RaceID
				}
			}
		}
		allRaces := tools.MonReturnAllRaces(tools.GetClient(), bson.M{"RaceID": getRace})
		for _, v := range allRaces {
			if v.RaceID == getRace {
				readyCount = v.PlayersReady
				joinCount = v.PlayersEntered
			}
		}
		if readyCount == joinCount {
			voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
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

		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Not all players are ready make sure everyone does .ready before starting the race.")
		}

		tools.MonUpdateRace(tools.GetClient(), bson.M{"Started":true, "Start Time":starttime}, bson.M{"RaceID":getRace})

	// You finish the Race
	case m.Content == ".done":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var timeJoined time.Time
		var started bool
		var playerDone bool
		var raceid string
		c := tools.GetClient()
		playerLookup := tools.MonReturnAllPlayers(c, bson.M{"PlayerID": m.Author.ID})
		fmt.Println(playerLookup)
		for _, v := range playerLookup {
			if v.PlayerID == m.Author.ID {
				timeJoined = v.JoinTime
				playerDone = v.Done
				raceid = v.RaceID
			}
		}
		lookupStart := tools.MonReturnAllRaces(tools.GetClient(), bson.M{"RaceID": raceid})
		for _, v := range lookupStart {
			if v.RaceID == raceid {
				started = v.Started
			}
		}
		var time2 = time.Now()
		var endtime1 = time2.Sub(timeJoined)
		var endtime = endtime1.Truncate(1 * time.Millisecond)

		if playerDone == true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You've already finished the race <@"+m.Author.ID+"> " +
				"Please wait for the next race.")
		} else if started == false {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">, Please wait for the race to start before finishing the race.")

		} else {
			tools.MonUpdatePlayer(c, bson.M{"Done Time": time2, "Done": true, "RaceID": "Done"}, bson.M{"PlayerID": m.Author.ID})
			_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
		}
	// Quit the race
	case m.Content == ".forfeit":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has forfeit the race, " +
			"hope you join us for the next race!")
		var playerRace string
		lookupPlayer := tools.MonReturnAllPlayers(tools.GetClient(), bson.M{"PlayerID": m.Author.ID})
		for _, v := range lookupPlayer {
			if v.PlayerID == m.Author.ID {
				if v.RaceID != "done" {
					playerRace = v.RaceID
				}
			}
		}
		lookupRace := tools.MonReturnAllRaces(tools.GetClient(), bson.M{"RaceID":playerRace})
		for _, v := range lookupRace {
			if v.RaceID == playerRace {
				enteredTotal := v.PlayersEntered
				readyTotal := v.PlayersReady
				fmt.Printf("%v, %v ", enteredTotal, readyTotal)

				tools.MonDeletePlayer(tools.GetClient(), bson.M{"PlayerID": m.Author.ID})
				tools.MonUpdateRace(tools.GetClient(), bson.M{"Players Entered": enteredTotal-1, "Players Ready": readyTotal-1}, bson.M{"RaceID": playerRace})
			}
		}
	}
}
