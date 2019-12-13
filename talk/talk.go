package talk

import (
	"donnybrook/tools"
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

	if strings.HasPrefix(m.Content, ".tts") {
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		text := strings.TrimPrefix(m.Content,".tts ")
		bytes, err := polly.Speech(text)
		if err != nil {
			panic(err)
		}
		tempdir, err := ioutil.TempDir("./media", "tts")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(tempdir+"message.mp3", bytes, 0644)
		if err != nil {
			panic(err)
		}
		voice, _ := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		tools.PlayAudioFile(voice, tempdir+"message.mp3")
		err = os.Remove(tempdir)
		if err != nil {
			panic(err)
		}
	}

}

