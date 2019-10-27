package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var db *sql.DB
var err error
var atime = time.Now()

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

	guildid := os.Getenv("DISCORD_GUILD")
	chanid := os.Getenv("DISCORD_VOICE")
	dgv, err := dg.ChannelVoiceJoin(guildid, chanid, false, true)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Playing Audio")

		files, _ := ioutil.ReadDir("./audio")
		for _, f := range files {
			fmt.Println("Play Audio File:", f.Name())
			dg.UpdateStatus(0, f.Name())
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s%s", "./audio", f.Name()), make(chan bool))
		}
	dgv.Close()

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

	// if m.Content == ".ready" {
	// s.ChannelMessageSend(m.ChannelID, m.Author.Username + " has joined the race.")
	// var sqlStatement = `
	// INSERT INTO races (name, starttime)
	// VALUES ($1, $2)`
	// _, err = db.Exec(sqlStatement, m.Author.Username, time.Now())
	// if err != nil {
	//	panic(err)
	// }
	// }
	switch {
	// Set up Race
	case m.Content == ".setup":
		var raceid = randomString(4)
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+", ready to start race "+raceid)
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
	case m.Content == ".forfit":
		s.ChannelMessageSend(m.ChannelID, m.Author.Username+" has forfeit the race, hope you join us for the next race!")
	}
}
