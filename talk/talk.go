package talk

import (
	"donnybrook/tools"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/leprosus/golang-tts"
	"io/ioutil"
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

	if strings.HasPrefix(m.Content, "v@echo") {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		text := strings.TrimPrefix(m.Content, "v@echo ")
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
		tools.PlayAudioFile(voice, tempFile.Name())
		err = os.Remove(tempFile.Name())
		if err != nil {
			panic(err)
		}
	}
}