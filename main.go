package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var staticuser = "108344508940316672"
var db *sql.DB
var err error
var atime = time.Now()
var slogan = "Donnybrook - Because sometimes fast needs to be quantified."

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dstoken := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + dstoken)
	if err != nil {
		log.Fatal("Error creating Discord session,", err)
		return
	}

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

	dg.Close()
}

func initDB() {

	dbhost := os.Getenv("PG_HOST")
	dbport := 5432
	dbuser := os.Getenv("PG_USER")
	dbpass := os.Getenv("PG_PASS")
	dbname := os.Getenv("PG_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbhost, dbport, dbuser, dbpass, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to database.")
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Set up Race
	if strings.HasPrefix(m.Content, ".setup") {
		var msgstr = m.Content
		msgstr = strings.TrimPrefix(msgstr, ".setup")
		msgstr = strings.TrimSpace(msgstr)
		msgary := strings.Split(msgstr, ",")
		if len(msgary) <= 1 {
			s.ChannelMessageSend(m.ChannelID, "You're missing either the game or the category of the race. run `.setup gamename, category`")
		} else {
			var arg1 = msgary[0]
			var arg2 = msgary[1]
			var raceid = randomString(4)
			racestring := fmt.Sprintf("%s, has started race %s for %s in category %s.", "<"+"@"+m.Author.ID+">", raceid, strings.TrimSpace(arg1), strings.TrimSpace(arg2))
			s.ChannelMessageSend(m.ChannelID, racestring)
		}
	}
	switch {
	// Join Race
	case m.Content == ".join":
		if m.Author.ID == staticuser {
			s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" You've already joined the race please wait for the race to start.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has joined the race.")
		}
	// Ready up for race
	case m.Content == ".ready":
		if m.Author.ID == staticuser {
			s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" You've already readied up for the race if you need to leave use `.unready`.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has readied up.")
		}
	// unready
	case m.Content == ".unready":
		s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has unreadied")
	// Start Race once all ready
	case m.Content == ".start":
		s.ChannelMessageSend(m.ChannelID, "All racers have readied")
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "Starting in 3.")
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "2.")
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "1.")
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, "Go!")
	// You finish the Race
	case m.Content == ".done":
		var time2 = time.Now()
		var endtime1 = time2.Sub(atime)
		var endtime = endtime1.Truncate(1 * time.Millisecond)
		s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has finished the race in: "+endtime.String())
	// Quit the race
	case m.Content == ".forfeit":
		s.ChannelMessageSend(m.ChannelID, "<"+"@"+m.Author.ID+">"+" has forfeit the race, hope you join us for the next race!")
	// Help text
	case m.Content == ".help":
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title: "Donnybrook Help:",
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128",
				Width:  128,
				Height: 128,
			},
			Description: "Welcome to the Donnybrook Race bot to get started use: `.setup` to create a race. \n" +
				"Example: ```.setup A Link to the Past, Any%``` \n" +
				"`.join` - Join the race. \n" +
				"`.ready` - Ready up for the race. \n " +
				"`.unready` - Unready from the race. \n" +
				"`.start` - Start the race once everyone is ready. \n" +
				"`.done` - You finished the race. \n" +
				"`.forfeit` - Leave the race early.\n" +
				"`.help` - You're reading it.",
			Color: 0x550000,
			Author: &discordgo.MessageEmbedAuthor{
				URL:     "https://donnybrookbot.xyz",
				Name:    "Donnybrook",
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=128"},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    slogan,
				IconURL: "https://cdn.discordapp.com/avatars/637392848307748895/7b5cb5a0cb148a5119a84f8a8201169f.png?size=16"}})
	}

}
