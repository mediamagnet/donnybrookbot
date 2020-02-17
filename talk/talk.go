package talk

import (
	"donnybrook/tools"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/leprosus/golang-tts"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"strconv"
	"strings"

	"log"

	"os"
)

func BotTalk(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := godotenv.Load()
	awsKeyID := os.Getenv("AWS_KEYID")
	awsSecID := os.Getenv("AWS_KEYSECRET")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	polly := golang_tts.New(awsKeyID, awsSecID)

	polly.Format(golang_tts.MP3)
	polly.Voice(golang_tts.Justin)

	switch {
	case strings.HasPrefix(m.Content, ".tts"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		text := strings.TrimPrefix(m.Content, ".tts ")
		bytes, err := polly.Speech(text)
		if err != nil {
			panic(err)
		}
		tempFile, err := ioutil.TempFile("./media", "tts*.mp3")
		fmt.Println(tempFile)
		if err != nil {
			panic(err)
		}
		fmt.Println(tempFile.Name())
		err = ioutil.WriteFile(tempFile.Name(), bytes, 0644)
		if err != nil {
			panic(err)
		}
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		tools.PlayAudioFile(voice, tempFile.Name(), m.GuildID)
		err = os.Remove(tempFile.Name())
		if err != nil {
			panic(err)
		}
	case strings.HasPrefix(m.Content, ".vol"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		volNew := strings.TrimPrefix(m.Content, ".vol")
		volNew1, _ := strconv.Atoi(volNew)

		if volNew1 >= 201 {
			bytes, err := polly.Speech("Sure but that might wreck your hearing.")
			if err != nil {
				panic(err)
			}
			tempFile, err := ioutil.TempFile("./media", "tts*.mp3")
			fmt.Println(tempFile)
			if err != nil {
				panic(err)
			}
			fmt.Println(tempFile.Name())
			err = ioutil.WriteFile(tempFile.Name(), bytes, 0644)
			if err != nil {
				panic(err)
			}
			voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
			tools.PlayAudioFile(voice, tempFile.Name(), m.GuildID)
		} else {
			settingsInput := tools.Settings{
				GuildID: m.GuildID,
				Volume:  volNew1,
			}
			volumeFound := 0
			volumeLookup := tools.MonReturnAllSettings(tools.GetClient(), bson.M{"GuildID": m.GuildID})
			for _, v := range volumeLookup {
				if v.GuildID == m.GuildID {
					volumeFound = v.Volume
				}
			}

			if volumeFound == 0 {
				tools.MonSettings("donnybrook", "settings", settingsInput)
			}
			tools.MonUpdateSettings(tools.GetClient(), bson.M{"Volume": volNew1}, bson.M{"GuildID": m.GuildID})
			bytes, err := polly.Speech("Volume has been set to " + volNew)
			if err != nil {
				panic(err)
			}
			tempFile, err := ioutil.TempFile("./media", "tts*.mp3")
			fmt.Println(tempFile)
			if err != nil {
				panic(err)
			}
			fmt.Println(tempFile.Name())
			err = ioutil.WriteFile(tempFile.Name(), bytes, 0644)
			if err != nil {
				panic(err)
			}
			voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
			tools.PlayAudioFile(voice, tempFile.Name(), m.GuildID)
		}

	}
}