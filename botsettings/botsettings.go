package botsettings

import (
	"donnybrook/tools"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
)

func setDefault(guildID string) {
	settingsDefault := tools.Settings{
		GuildID: guildID,
		CanTalk: true,
		BotFun:  true,
	}
	tools.MonSettings("donnybrook", "settings", settingsDefault)
}

var SettingCanTalk = true

func BotSettings(s *discordgo.Session, m *discordgo.MessageCreate) {
	switch {

	case m.Content == "s.default":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		_, _ = s.ChannelMessageSend(m.ChannelID, "Setting default")
		setDefault(m.GuildID)

	case m.Content == "s.talking":
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
		SettingGuildID := ""
		c := tools.GetClient()
		settingLookUp := tools.MonReturnAllSettings(c, bson.M{"GuildID": m.GuildID})
		fmt.Println(settingLookUp)
		for _, v := range settingLookUp {
			if v.GuildID == m.GuildID {
				SettingCanTalk = v.CanTalk
				SettingGuildID = v.GuildID
			}
		}
		if SettingCanTalk == true {
			SettingCanTalk = false
			tools.MonUpdateSettings(c, bson.M{"Can Talk": SettingCanTalk}, bson.M{"GuildID": SettingGuildID})
			_, _ = s.ChannelMessageSend(m.ChannelID, "Okay, turned off")
		} else {
			SettingCanTalk = true
			tools.MonUpdateSettings(c, bson.M{"Can Talk": SettingCanTalk}, bson.M{"GuildID": SettingGuildID})
			_, _ = s.ChannelMessageSend(m.ChannelID, "Okay, turned on... are you?")
		}

	}
}
