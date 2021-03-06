package main

import (
	"donnybrook/botadmin"
	"donnybrook/botmain"
	"donnybrook/racebot"
	"donnybrook/talk"
	"donnybrook/tools"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)


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

	dg.AddHandler(connect)

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

func connect(s *discordgo.Session, c *discordgo.Connect) {
	fmt.Println(c)
	var guildName = make([]string, 1)
	for _, v := range s.State.Guilds {
		guildName = append(guildName, v.Name)
	}
	for {
		guildCount := len(guildName)-1
		err := s.UpdateListeningStatus(fmt.Sprintf("races in %v servers", guildCount))
		time.Sleep(15 * time.Minute)
		err = s.UpdateListeningStatus("cosmic background radiation")
		time.Sleep(15 * time.Minute)
		err = s.UpdateStatus(0, "Donnybrook v0.0.1")
		time.Sleep(15 * time.Minute)
		err = s.UpdateListeningStatus(".help")
		time.Sleep(15 * time.Minute)
		err = s.UpdateStatus(0, "https://donnybrookbot.xyz")
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(15 * time.Minute)
	}

}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Test to make sure bot isn't talking to self.
	if m.Author.ID == s.State.User.ID {
		return
	}
	if tools.ComesFromDM(s, m) == false {
		racebot.RaceBot(s, m)
		botadmin.BotAdmin(s, m)
		talk.BotTalk(s, m)
		botmain.BotMain(s, m)

		tools.RandomString(4)
	} else if tools.ComesFromDM(s, m) == true && m.Author.ID == "108344508940316672" {
		switch {
		case m.Content == ".lick":
			_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

			//voice, _ := tools.JoinUserVoiceChannel(s, "324233923506733058", "108344508940316672", "108344598018957312")
			voice, _ := s.ChannelVoiceJoin("108344598018957312", "324233923506733058", false, false)
			go s.ChannelMessageSend(m.ChannelID, "https://i.ytimg.com/vi/lSXxEdaOqgU/maxresdefault.jpg")
			go tools.PlayAudioFile(voice, "media/lick.mp3", "108344598018957312", false)
			time.Sleep(50 * time.Second)
			_, _ = s.ChannelVoiceJoin("108344598018957312", "", false, false)
		case strings.HasPrefix(m.Content, ".asay"):
			msg := strings.TrimPrefix(m.Content, ".asay ")
			msg1 := strings.Split(msg, ",")
			channel := msg1[0]
			body := msg1[1]
			s.ChannelMessageSend(channel, body)

		}

	} else {
		fmt.Println("It's a DM")
	}
}
