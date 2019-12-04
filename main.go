package main

import (
	"donnybrook/botadmin"
	"donnybrook/botsettings"
	"donnybrook/tools"
	"donnybrook/racebot"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var slogan = "Donnybrook - Because sometimes fast needs to be quantified."
var startTime1 = time.Now()
// var endTime1 = time.Now()
var wg sync.WaitGroup
var SettingCanTalk = true

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

	racebot.RaceBot(s, m)
	botadmin.BotAdmin(s, m)
	botsettings.BotSettings(s, m)
	fmt.Printf("%20s %20s %20s > %s\n", m.ChannelID, time.Now().Format(time.Stamp), m.Author.Username, m.Content)
	tools.RandomString(4)

	switch {
	// Ready up for race
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
		if SettingCanTalk == true {
			_, err := tools.JoinUserVoiceChannel(s,m.ChannelID, m.Author.ID)
			if err != nil {
				fmt.Println(err)
				s.ChannelMessageSend(m.ChannelID, "Error: Try joining a voice channel first.")
			}
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry can't talk right now.")
		}
	// Leave voice
	case m.Content == "v.leave":
		// _ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

		// Clean up channel currently in.

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
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID)
		go s.ChannelMessageSend(m.ChannelID, "Peace was never an option \n " +
			"https://i.kym-cdn.com/photos/images/newsfeed/001/597/651/360.jpg")
		go tools.PlayAudioFile(voice, "media/honk.mp3")
		time.Sleep(10 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	case m.Content == ".lick":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID)
		go s.ChannelMessageSend(m.ChannelID, "https://i.ytimg.com/vi/lSXxEdaOqgU/maxresdefault.jpg")
		go tools.PlayAudioFile(voice, "media/lick.mp3")
		time.Sleep(50 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	}
}