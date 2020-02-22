package botmain

import (
	"donnybrook/tools"
	"github.com/bwmarrin/discordgo"
	"time"
	"fmt"
)

var slogan = "Donnybrook - Because sometimes fast needs to be quantified."
var launchTime = time.Now()
// var endTime1 = time.Now()

func BotMain(s *discordgo.Session, m *discordgo.MessageCreate) {
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
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=128"},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=128",
				Width:  128,
				Height: 128,
			},
			Color:       0x550000,
			Description: "Welcome to the Donnybrook Race bot, Here's some useful commands: \n",
			Fields: []*discordgo.MessageEmbedField{
				{Name: ".setup Game, Category", Value: "Setup a race"},
				{Name: ".join <race id>", Value: "Join a race with the specified id"},
				{Name: ".ready", Value: "Ready up for the race after you're done setting up"},
				{Name: ".unready", Value: "Leave the ready state for the race."},
				{Name: ".start", Value: "Once everyone is ready start the race."},
				{Name: ".done", Value: "Once you finish the race use this."},
				{Name: ".forfeit", Value: "Not able to keep going use this to exit the race."},
				{Name: ".help", Value: "You're reading it."},
				{Name: ".invite", Value: "Got a question? join the Donnybrook discord here: https://discord.gg/cyZzPZY"},
				{Name: ".voicejoin", Value: "Join a voice channel for voiced countdown. Must be in voice channel for this to work."},
				{Name: ".voiceleave", Value: "Leave the voice channel"},
				{Name: ".tts", Value: "Say something in your voice chat"},
				{Name: ".cleanup", Value: "Cleans messages in channel command is run in. User must have `Manage Message` permissions."},
				{Name: ".scatter <Channels to scatter to>", Value: "Scatter users to the provided channels. User must have `Manage Server` Permissions."},
				{Name: ".lick", Value: "..."}},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    slogan,
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=16"}})
	// Helf text
	case m.Content == ".helf":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "Donnybrook Helf:",
			Author: &discordgo.MessageEmbedAuthor{
				URL:     "https://donnybrookbot.xyz",
				Name:    "Donnybrook",
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=128"},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=128",
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
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/b20c5e4c7ac3587a7c2438a78722b425.png?size=16"}})
	// Join voice
	case m.Content == ".voicejoin":

		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voiceState, err := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(voiceState)

	// Leave voice
	case m.Content == ".voiceleave":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	case m.Content == ".uptime":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var uptime1 = time.Now()
		var uptime = uptime1.Sub(launchTime)
		uptime = uptime.Truncate(1 * time.Millisecond)
		upTimeLine := fmt.Sprintf("I've been running for: %v", uptime)
		err, upTimed := s.ChannelMessageSend(m.ChannelID, upTimeLine)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(upTimed)
	// Bees
	case m.Content == ".invite":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Got a question? join the Donnybrook discord here: https://discord.gg/cyZzPZY")
	case m.Content == "b.swarm":
		lastMsg := make([]string, 1)
		fmt.Println(lastMsg)
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "\n :bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n "+
			":bee: :bee: :bee: :bee: \n ")
	case m.Content == ".honk":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, err := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: you must be in a voice channel first.")
		}
		go s.ChannelMessageSend(m.ChannelID, "Peace was never an option \n "+
			"https://i.kym-cdn.com/photos/images/newsfeed/001/597/651/360.jpg")
		go tools.PlayAudioFile(voice, "media/honk.mp3", m.GuildID, false)
		time.Sleep(10 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	case m.Content == ".lick":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		go s.ChannelMessageSend(m.ChannelID, "https://i.ytimg.com/vi/lSXxEdaOqgU/maxresdefault.jpg")
		go tools.PlayAudioFile(voice, "media/lick.mp3", m.GuildID, false)
		time.Sleep(50 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false, false)

	case m.Content == ".terminate":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		go s.ChannelMessageSend(m.ChannelID, "I'll be back.")
		go tools.PlayAudioFile(voice, "media/DEDENDENDEDEN.mp3", m.GuildID, false)
		time.Sleep(45 * time.Second)
		_, _ = s.ChannelVoiceJoin(m.GuildID, "", false,false)

	}
}
