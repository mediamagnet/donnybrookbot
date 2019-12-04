package botadmin

import (
	"donnybrook/tools"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func BotAdmin(s *discordgo.Session, m *discordgo.MessageCreate) {
	switch {
	case strings.HasPrefix(m.Content, "a.scatter"):
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		var voice *discordgo.VoiceConnection
		voice, _ = tools.JoinUserVoiceChannel(s, m.ChannelID, m.Author.ID)
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
	case m.Content == "a.cleanup":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

		if tools.MemberHasPermission(s, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages|discordgo.PermissionAdministrator) {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Deleting messages in <#"+m.ChannelID+">")
			time.Sleep(3 * time.Second)
			wg.Add(100)
			fmt.Println("Cleanup Requested")
			for i := 0; i < 1000; i++ {
				messages, _ := s.ChannelMessages(m.ChannelID, 1, "", "", "")
				if len(messages) == 0 {
					break
				}
				fmt.Println(messages[0].ID)
				time.Sleep(500 * time.Millisecond)
				_ = s.ChannelMessageDelete(m.ChannelID, messages[0].ID)
				fmt.Println("done cleaning", i)
			}
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Sorry <@"+m.Author.ID+"> You need Manage Message permissions to run a.cleanup")
		}
	}
}
