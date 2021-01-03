package botadmin

import (
	"donnybrook/tools"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var wg sync.WaitGroup

// BotAdmin commands
func BotAdmin(s *discordgo.Session, m *discordgo.MessageCreate) {
	switch {
	case strings.HasPrefix(m.Content, ".scatter"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var voice *discordgo.VoiceConnection
		voice, err := tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID, m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: you must be in a voice channel first.")
		}
		var msgstr = m.Content
		var msgarr = make([]string, 1)
		msgstr = strings.TrimPrefix(msgstr, ".scatter ")
		msgarr = strings.Split(msgstr, ",")
		fmt.Println(msgarr)
		rand.Seed(time.Now().Unix())
		vUsers := tools.FindAllVoiceState(s)
		if tools.MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageServer|discordgo.PermissionAll) {
			wg.Add(len(vUsers))
			// vChan := voiceChannels(s, m.GuildID)
			tools.PlayAudioFile(voice, "media/scatter2.mp3", m.GuildID, false)
			for i := 0; i <= len(vUsers)-1; i++ {
				go func(i int) {
					defer wg.Done()
					vChan := rand.Intn(len(msgarr))
					pickedChan := tools.ChannelIDFromName(s, m.GuildID, msgarr[vChan])
					choiceUser := vUsers[i]
					fmt.Println(pickedChan)
					vChan = rand.Intn(len(msgarr))
					pickedChan = tools.ChannelIDFromName(s, m.GuildID, msgarr[vChan])

					_ = s.GuildMemberMove(m.GuildID, choiceUser, pickedChan.Channel)
					time.Sleep(500 * time.Millisecond)
				}(i)
			}
			wg.Wait()
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "No <@"+m.Author.ID+"> you cannot scatter people.")
		}
	case strings.HasPrefix(m.Content, ".cleanup"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		delCount1 := strings.TrimPrefix(m.Content, ".cleanup ")
		delCount, _ := strconv.Atoi(delCount1)
		fmt.Println(delCount)

		if tools.MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Deleting messages in <#"+m.ChannelID+">")
			if delCount == 0 {
				time.Sleep(3 * time.Second)
				wg.Add(100)
				fmt.Println("Cleanup Requested")
				for i := 0; i < 1000; i++ {
					messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
					if len(messages) == 0 {
						break
					}
					// fmt.Println(messages[0].ID)
					println(i)
					time.Sleep(500 * time.Millisecond)
					_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
					fmt.Println("done cleaning", i)
				}
			} else {
				time.Sleep(3 * time.Second)
				wg.Add(100)
				fmt.Println("Cleanup Requested")
				for i := 0; i < delCount+1; i++ {
					messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
					if len(messages) == 0 {
						break
					}
					fmt.Println(messages[0].ID)
					time.Sleep(500 * time.Millisecond)
					_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
					fmt.Println("done cleaning", i)
				}
			}
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+"> You need Manage Message permissions to run .cleanup")
		}
	case strings.HasPrefix(m.Content, ".void "):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

		if strings.Contains(strings.TrimPrefix(m.Content, ".void"), "start") {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Be careful as the void might yell back.")
			_, _ = s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
				Topic:    "Be careful as the void might yell back.",
				Position: 99,
			})
			tools.MonUpdateSettings(tools.GetClient(), bson.M{"VoidChan": m.ChannelID}, bson.M{"GuildID": m.GuildID})
			for {
				wg.Add(1)
				mCount, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
				if len(mCount) == 0 {
					time.Sleep(30 * time.Second)
				} else if len(mCount) >= 1 {
					for i := 0; i < 1000; i++ {
						messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
						if len(messages) == 0 {
							break
						}
						time.Sleep(500 * time.Millisecond)
						_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
						fmt.Println("done cleaning", i)
					}
				}
				wg.Done()
			}
		} else if strings.Contains(strings.TrimPrefix(m.Content, ".void "), "stop") {
			tools.MonDeleteSettings(tools.GetClient(), bson.M{"VoidChan": m.ChannelID})
			s.ChannelMessageSend(m.ChannelID, "The Void sleeps for now.")
			_, _ = s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
				Topic:    " ",
				Position: 99,
			})
		}
	}
}
