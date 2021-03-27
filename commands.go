package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Ping command which returns a message
func Ping(ctx *Context) {
	m := ctx.NewEmbed("Pinging....")
	ts, _ := m.Timestamp.Parse()
	now := time.Now()
	ctx.EditEmbed(m, fmt.Sprintf("🏓 | **Pong my ping**\n\n💗 | **Heartbeat**: `%1.fms`\n ⏱️| **Message Delay**: `%1.fms`",
		float64(ctx.Session.HeartbeatLatency().Milliseconds()),
		math.Abs(float64(now.Sub(ts).Milliseconds()))))
}

// Help returns a list of commands and their help
func Help(ctx *Context) {
	args := strings.Split(strings.ToLower(ctx.Msg.Content), " ")[1:]
	helpMsg := ""

	cmds := UniqueCmds(Commands)

	for _, c := range cmds {
		helpMsg += fmt.Sprintf("```css\n%s%s - %s\n```", Config.Prefix, c.Name, c.Desc[0])
	}

	if len(args) < 1 {
		ctx.NewEmbed(fmt.Sprintf("📚 | **%d** Total Commands\n%s\nUse **%shelp [COMMAND]** for more info on a command.", len(cmds), helpMsg, Config.Prefix))
		return
	}
	cmd, ok := Commands[args[0]]

	if !ok {
		ctx.SendErr("no command found")
		return
	}

	ctx.SendCommandHelp(cmd)
}

// Find finds a match between users
func Find(ctx *Context) {
	var matches struct {
		Likes []string
		Match []*DBUser
	}

	db := *NewDB()
	userDB, ok := db[ctx.Msg.Author.ID]
	if !ok {
		ctx.SendErr(fmt.Sprintf("opt in to be able to find. (use %sopt)", Config.Prefix))
		return
	}

	for _, v := range db {
		// Ignore same user or different guildIDs
		if v.UserID == ctx.Msg.Author.ID || v.GuildID != ctx.Msg.GuildID {
			continue
		}

		// Tags are similar, add them to matches
		for _, tag := range v.Tags {
			ok, t := Contains(userDB.Tags, tag)
			if ok {
				matches.Match = append(matches.Match, v)
				matches.Likes = append(matches.Likes, t)
			}
		}
	}

	// Remove duplicates
	matches.Match = UniqueDBUser(matches.Match)

	// No matches found
	if len(matches.Match) <= 0 {
		ctx.NewEmbed("🇫 | **No matches** found matching your interests! Maybe try again later?")
		return
	}

	// DM User matches for them to choose from
	channel, err := ctx.Session.UserChannelCreate(ctx.Msg.Author.ID)
	u, err := ctx.Session.GuildMember(matches.Match[0].GuildID, matches.Match[0].UserID)
	t, err := u.JoinedAt.Parse()
	if err != nil {
		ctx.SendErr(err)
		return
	}

	for i, t := range db[u.User.ID].Tags {
		for _, t2 := range matches.Likes {
			if t == t2 {
				db[u.User.ID].Tags[i] = fmt.Sprintf("`%s`", db[u.User.ID].Tags[i])
			}
		}
	}

	embedToSend := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: fmt.Sprintf("👤 | This is **[** `%s#%s` **]**, what do you think?", u.User.Username, u.User.Discriminator),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "📖 Join Date",
					Value:  fmt.Sprintf("**%d/%d/%d**", t.Month(), t.Day(), t.Year()),
					Inline: true,
				},
				{
					Name:   "💬 Interests",
					Value:  fmt.Sprintf("**%s**", strings.Join(db[u.User.ID].Tags, ", ")),
					Inline: true,
				},
			},
			Image: &discordgo.MessageEmbedImage{URL: u.User.AvatarURL("")},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("1/%d Matches", len(matches.Match)),
			},
		},
	}
	// Send message
	ctx.NewEmbed(fmt.Sprintf("📬 | Found **%d total matches**, check your DMs!", len(matches.Match)))
	ctx.Session.ChannelMessageSendComplex(channel.ID, embedToSend)
	collector :=  &MessageCollector{
		MessagesCollected: []*discordgo.Message{},
		Filter:            func(ctx *Context, m *discordgo.Message) bool {
			if m.Author.ID != ctx.Msg.Author.ID {
				return false
			}
			return true
		},
		EndAfterOne:       true,
		UseTimeout:        true,
		Timeout: time.Minute,
	}
	ctx.Session.ChannelMessageSend(channel.ID, "```css\nEnter your response:\n\nYes\nNo\nCancel\n```")

	err = collector.New(ctx)
	if err != nil {
		ctx.SendErr(err)
		return
	}

	switch strings.ToLower(collector.MessagesCollected[0].Content) {
	case "yes":
		// Create private channel in guild
		ch, err := ctx.Session.GuildChannelCreate(ctx.Msg.GuildID, fmt.Sprintf("%s-and-%s", ctx.Msg.Author.Username, u.User.Username), discordgo.ChannelTypeGuildText)
		if err != nil {
			ctx.SendErr(err)
			return
		}
		ctx.Session.ChannelMessageSend(ch.ID, fmt.Sprintf("<@%s> and <@%s>, welcome!", ctx.Msg.Author.ID, u.User.ID))

	case "no":
		// TODO: Find next match
		return
	}


}

// Opt register the user for the service
func Opt(ctx *Context) {
	m := strings.Join(strings.Split(strings.ToLower(ctx.Msg.Content), " ")[1:], " ")

	if len(m) < 1 {
		ctx.SendCommandHelp(ctx.Command)
		return
	}

	args := strings.Split(m, ",")

	// Remove them from the service
	if args[0] == "out" {
		err := RemoveUserFromDatabase(ctx.Msg.Author.ID)

		if err != nil {
			ctx.SendErr(err)
			return
		}

		ctx.NewEmbed(":x: | Done, you are **now removed** from the service!")
		return
	}

	// Remove duplicates
	args = Unique(args)

	var parsedArgs []string
	for i, a := range args {
		if len(a) < 1 {
			ctx.SendErr(fmt.Sprintf("Tag #%d is too short, it should be >1 character.", i+1))
			return
		}

		// Check if first character is a space, if so remove it
		if a[0] == ' ' {
			parsedArgs = append(parsedArgs, a[1:])
			continue
		}

		if len(a) >= 50 {
			ctx.SendErr(fmt.Sprintf("Tag #%d is too long, it should be <50 characters.", i+1))
			return
		}

		parsedArgs = append(parsedArgs, a)
	}

	// Add them to the database
	dbUser := &DBUser{
		UserID:       ctx.Msg.Author.ID,
		Tags:         parsedArgs,
		Occupied:     false,
		CurrentMatch: "",
		History:      nil,
		TimeStart:    time.Now(),
		GuildID:      ctx.Msg.GuildID,
	}
	AddUserToDatabase(dbUser)

	ctx.NewEmbed(fmt.Sprintf("✅ | Done, you are **now added** to the service!\n```css\nTags: %s\n```", strings.Join(parsedArgs, ",")))
}
