package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

var LastMessage = make(chan *discordgo.Message)
var LastReaction = make(chan *discordgo.MessageReaction)

// MessageCreate the function which handles message events
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bots
	if m.Author.Bot {
		return
	}
	// Ignore no content messages
	if m.Content == "" {
		return
	}

	// Create new context on each message1
	msg := strings.Split(strings.ToLower(m.Message.Content)[1:], " ")

	c := msg[0]
	// Find the command with the matching name alias and run it
	cmd, ok := Commands[c]
	if !ok {
		go func() { LastMessage <- m.Message }()
		return
	}
	// Make sure message starts with prefix
	if string(m.Message.Content[0]) != Config.Prefix {
		return
	}
	// Make sure it's in guild
	channel, _ := s.Channel(m.ChannelID)
	if channel.Type == discordgo.ChannelTypeDM {
		return
	}

	ctx := &Context{
		Session:      s,
		Msg:          m,
		Command:      cmd,
		LastMessage:  LastMessage,
		LastReaction: LastReaction,
	}
	cmd.Handler(ctx)
}

// MessageReactionAdd the function which handles message reaction events
func MessageReactionAdd(_ *discordgo.Session, m *discordgo.MessageReactionAdd) {
	LastReaction <- m.MessageReaction
}

// Ready the function which handles when the bot is ready
func Ready(s *discordgo.Session, e *discordgo.Ready) {
	fmt.Printf("[%s] - in as %s%s with prefix: \"%s\"\n", Purple.Sprint("BOT"), Yellow.Sprint(e.User.Username), Yellow.Sprint("#"+e.User.Discriminator), Green.Sprint(Config.Prefix))
	err := s.UpdateGameStatus(0, fmt.Sprintf("Eating Cheetos | %shelp", Config.Prefix))
	Die(err)
}
