package bot

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zmb3/spotify"
)

var BotToken string

const redirectURI = "http://localhost:8080/callback"

var html = `
<br/>
<a href="/player/play">Play</a><br/>
<a href="/player/pause">Pause</a><br/>
<a href="/player/next">Next track</a><br/>
<a href="/player/previous">Previous Track</a><br/>
<a href="/player/shuffle">Shuffle</a><br/>
`

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

var clientID = "your_client_id"
var clientSecret = "your_client_secret"

var client *spotify.Client
var playerState *spotify.PlayerState

func Run() {

	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	discord.AddHandler(newMessage)

	http.HandleFunc("/callback", completeAuth)

	http.HandleFunc("/player/", handlePlayer)

	http.HandleFunc("/", handleDefualt)

	discord.Open()

	defer discord.Close()

	go spotifyRoutine()

	http.ListenAndServe(":8080", nil)

}

func spotifyRoutine() {
	auth.SetAuthInfo(clientID, clientSecret)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	client = <-ch

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	playerState, err = client.PlayerState()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found your %s (%s)\n", playerState.Device.Type, playerState.Device.Name)
}

func handleDefualt(w http.ResponseWriter, r *http.Request) {
	log.Println("Got request for:", r.URL.String())
}

func handlePlayer(w http.ResponseWriter, r *http.Request) {
	// Your existing player handling logic
	action := strings.TrimPrefix(r.URL.Path, "/player/")
	fmt.Println("Got request for:", action)
	var err error
	switch action {
	case "play":
		err = client.Play()
	case "pause":
		err = client.Pause()
	case "next":
		err = client.Next()
	case "previous":
		err = client.Previous()
	case "shuffle":
		playerState.ShuffleState = !playerState.ShuffleState
		err = client.Shuffle(playerState.ShuffleState)
	}
	if err != nil {
		log.Print(err)
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {

	if message.Author.ID == discord.State.User.ID {
		return
	}
	log.Println("Got request for:", message.Content)

	var err error
	switch {
	case strings.Contains(message.Content, "!gomessage"):
		discord.ChannelMessageSend(message.ChannelID, "Hi")
	case strings.Contains(message.Content, "!goplay"):
		err = client.Play()
		discord.ChannelMessageSend(message.ChannelID, "playing")
	case strings.Contains(message.Content, "!gopause"):
		err = client.Pause()
		discord.ChannelMessageSend(message.ChannelID, "pausing")
	case strings.Contains(message.Content, "!gonext"):
		err = client.Next()
		discord.ChannelMessageSend(message.ChannelID, "playing next song")
	case strings.Contains(message.Content, "!goprev"):
		err = client.Previous()
		discord.ChannelMessageSend(message.ChannelID, "playing previous song")
	case strings.Contains(message.Content, "!goshuffle"):
		playerState.ShuffleState = !playerState.ShuffleState
		err = client.Shuffle(playerState.ShuffleState)
		discord.ChannelMessageSend(message.ChannelID, "shuffling")
	}
	if err != nil {
		log.Print(err)
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!"+html)
	ch <- &client
}
