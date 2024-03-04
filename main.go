package main

import bot "example.com/spotify_bot/bot"

func main() {
	bot.BotToken = "your_discord_bot_token"
	bot.Run() // call the run function of bot/bot.go
}
