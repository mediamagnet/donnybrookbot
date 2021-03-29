package tools

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/jonas747/dca"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"math/rand"
	"time"
)

//CList blah
type CList struct {
	Channel *string
}

var err error


func VoiceChannels(s *discordgo.Session, guildID string) []string {
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

func ChannelIDFromName(s *discordgo.Session, guildID string, channelName string) CList {
	channels, _ := s.GuildChannels(guildID)
	var c = ""
	for _, cList1 := range channels {
		if cList1.Name == channelName {
			c = cList1.ID
		}
	}
	return CList{
		Channel: &c,
	}
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(RandomInt(65, 90))
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

func FindUserVoiceState(session *discordgo.Session, userID string, guildID string) (*discordgo.VoiceState, error) {
	guild, err := session.State.Guild(guildID)
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}
	if err != nil {
		fmt.Println("Could not find guild specified")
	}

	return nil, errors.New("could not find user's voice state")
}

func FindAllVoiceState(session *discordgo.Session) []string {
	vString := make([]string, 1)
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			vString = append(vString, vs.UserID)
		}
	}
	return vString
}

func CurrentVoiceChannel(session *discordgo.Session, userID string) []string {
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

func JoinUserVoiceChannel(session *discordgo.Session, channelID string, userID string, guildID string) (*discordgo.VoiceConnection, error) {
	// Find a user's current voice channel
	vs, err := FindUserVoiceState(session, userID, guildID)
	if err != nil {
		fmt.Println(err)
	}
	//
	if vs == nil {
		session.ChannelMessageSend(channelID, "You're not in a Voice Channel.")
	} else {
		return session.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, false)
	}
	return nil, err
}

func PlayAudioFile(v *discordgo.VoiceConnection, filename string, guildID string, queued bool) {

	// Send "speaking" packet over the voice websocket

	err := v.Speaking(true)
	if err != nil {
		log.Fatal("Failed setting speaking", err)
	}
	volumeFound := 100

	volumeLookup := MonReturnAllRecords(GetClient(), bson.M{"GuildID": guildID}, "donnybrook", "Settings")
	for _, v := range volumeLookup {
		if v.Players.GuildID == guildID {
			volumeFound = v.Settings.Volume
		}
	}

	// Send not "speaking" packet over the websocket when we finish
	defer v.Speaking(false)

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120
	opts.Volume = volumeFound

	if queued == false {
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
	} else if queued == true {
		queue := goconcurrentqueue.NewFIFO()
		queue.Enqueue(filename)

		if queue.GetLen() == 0 {
			fmt.Println("Queue Empty")
		} else {
			for i := 0; i <= queue.GetLen()-1; i++ {
				item, err := queue.Dequeue()
				if err != nil {
					fmt.Println(err)
					return
				}

				item1 := fmt.Sprintf("%v", item)

				encodeSession, err := dca.EncodeFile(item1, opts)
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
		}

	}
}

func ComesFromDM(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false
		}
	}

	return channel.Type == discordgo.ChannelTypeDM
}
