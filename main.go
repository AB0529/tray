package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Commands all the commands for the bot
var (
	Commands = make(map[string]*Command)
	Config   *Conf
)

func main() {
	// Load config
	Config = NewConfig("./config.yml")
	// Setup Discord
	dg, _ := discordgo.New("Bot " + Config.Token)
	// Register events
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	dg.AddHandler(Ready)
	dg.AddHandler(MessageCreate)
	dg.AddHandler(MessageReactionAdd)

	// Register commands
	RegisterCommands([]*Command{
		{
			Name:    "ping",
			Aliases: []string{"pong"},
			Example: []string{Config.Prefix + "ping"},
			Desc:    []string{"Generic Ping-Pong command"},
			Handler: Ping,
		},
		{
			Name:    "opt",
			Aliases: []string{"o"},
			Example: []string{Config.Prefix + "opt sports, anime", Config.Prefix + "opt out"},
			Desc:    []string{"Registers the user for the service or removes them"},
			Handler: Opt,
		},
		{
			Name:    "find",
			Aliases: []string{"f"},
			Example: []string{Config.Prefix + "find"},
			Desc:    []string{"Finds users with similar interests"},
			Handler: Find,
		},
		{
			Name:    "help",
			Aliases: []string{"h"},
			Example: []string{Config.Prefix + "help", Config.Prefix + "help opt"},
			Desc:    []string{"Gives you info on commands and lists all the commands"},
			Handler: Help,
		},
	})

	// Open a websocket connection to Discord and begin listening.
	err := dg.Open()
	if err != nil {
		Die("could not creating session")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	err = dg.Close()
	Die(err)
}
