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

	"github.com/Necroforger/dgwidgets"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var err error
var atime = time.Now()
var slogan = "Donnybrook - Because sometimes fast needs to be quantified."

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("PG_HOST")
	port := 5432
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASS")
	dbname := os.Getenv("PG_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
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
		fmt.Printf("%v", msgary)
		var arg1 = msgary[0]
		var arg2 = msgary[1]
		var raceid = randomString(4)
		racestring := fmt.Sprintf("%s, has started race %s for %s in category %s.", m.Author.Username, raceid, arg1, arg2)
		s.ChannelMessageSend(m.ChannelID, racestring)
	}
	switch {
	// Join Race
	case m.Content == ".join":
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has joined the race.")

	// Ready up for race
	case m.Content == ".ready":
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has readied up.")
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
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has finished the race in: "+endtime.String())
	// Quit the race
	case m.Content == ".forfeit":
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has forfeit the race, hope you join us for the next race!")
	// Help text
	case m.Content == ".help":
		p := dgwidgets.NewPaginator(s, m.ChannelID)
		p.Add(
			&discordgo.MessageEmbed{
				Title:       "Donnybrook Help:",
				Description: "Welcome to the Donnybrook Race bot to get started use: `.setup` to create a race. \n Example: ```.setup A Link to the Past, Any%``` \n The rest of the commands are on page two.",
				Color:       0x550000},
			&discordgo.MessageEmbed{
				Title:       "Donnybrook Help",
				Description: ".join - Join the race. \n .ready - Ready up for the race. \n .start - Start the race once everyone is ready. \n .done - You finished the race. \n .forfeit - Leave the race early. \n .help - You're reading it.",
				Color:       0x550000},
			&discordgo.MessageEmbed{
				Title:       "Page three",
				Description: "Oh look you're on page three \n https://i.imgur.com/d86T9Mw.png",
				Color:       0x550000})
		p.SetPageFooters()

		p.ColourWhenDone = 0xffff

		p.Widget.Timeout = time.Minute * 5

		p.Spawn()

	case m.Content == ".save":
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" I don't know how to tell you this but you're in the matrix, It's time to wake up.")

	case m.Content == ".wake":
		s.ChannelMessageSend(m.ChannelID, "Error Invalid command: Can't wake up.")
	}

}
