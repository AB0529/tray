package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/color"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

var (
	// Red color red
	Red = color.Red
	// Purple color purple
	Purple = color.Magenta
	// Green color green
	Green = color.LightGreen
	// Yellow color yellow
	Yellow = color.Yellow
)

// Warn logs warning to stdout
func Warn(err interface{}) {
	// Handle strings being passed by creating an error type
	if err != nil {
		if e, ok := err.(string); ok {
			err = errors.New(e)
		}
		fmt.Printf("[%s] - %v\n", Yellow.Sprint("WARN"), err)
	}
}

// Die handles errors and exits
func Die(err interface{}) {
	// Handle strings being passed by creating an error type
	if err != nil {
		if e, ok := err.(string); ok {
			err = errors.New(e)
		}
		fmt.Printf("[%s] - %v\n", Red.Sprint("ERR"), err)
		os.Exit(1)
	}
}

// NewConfig creates the config
func NewConfig(path string) *Conf {
	var cfg *Conf
	d, _ := ioutil.ReadFile("./config.yml")
	yaml.Unmarshal(d, &cfg)

	return cfg
}

// RegisterCommands register provided commands
func RegisterCommands(cmds []*Command) {
	// Loop through each command and add it to Commands slice
	for _, c := range cmds {
		Commands[c.Name] = c
		for _, a := range c.Aliases {
			Commands[a] = c
		}
		fmt.Printf("[%s] - Loaded %s command\n", Purple.Sprint("CMD"), Yellow.Sprint(c.Name))
	}
}

// NewEmbed creates a simple embed with description only and a random color
func (ctx *Context) NewEmbed(content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageSendComplex(ctx.Msg.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: content,
			Footer:      &discordgo.MessageEmbedFooter{IconURL: ctx.Msg.Author.AvatarURL("")},
		},
	})
	Warn(err)
	return m
}

// Send sends a message to the same channel as the command
func (ctx *Context) Send(content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, content)
	Warn(err)
	return m
}

// SendErr sends a error message to the same channel as the command
func (ctx *Context) SendErr(content interface{}) {
	// Handle strings being passed by creating an error type
	if content != nil {
		if e, ok := content.(string); ok {
			content = errors.New(e)
		}
		ctx.NewEmbed(fmt.Sprintf(":x: | Uh oh, something **went wrong**!\n```css\n%s\n```", content))
	}
}

// Edit edits a message with a new content
func (ctx *Context) Edit(m *discordgo.Message, content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEdit(m.ChannelID, m.ID, content)
	Warn(err)
	return m
}

// EditEmbed edits a embed with a new content
func (ctx *Context) EditEmbed(m *discordgo.Message, content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: content,
			Footer:      &discordgo.MessageEmbedFooter{IconURL: ctx.Msg.Author.AvatarURL("512x512")},
		},
		ID:      m.ID,
		Channel: ctx.Msg.ChannelID,
	})
	Warn(err)
	return m
}

// FindCommandFlag finds a flag and the value in a command string
func (ctx *Context) FindCommandFlag(d string) ([]*Flag, error) {
	var foundFlags []*Flag

	// Arguments separated by a space
	args := strings.Split(strings.ToLower(ctx.Msg.Message.Content), d)[1:]

	if len(args) <= 0 {
		return nil, errors.New("no flags provided")
	}

	// Find the flag with the same name as args
	for i, arg := range args {
		for _, flag := range ctx.Command.Flags {
			if arg == flag.Name {
				// Add flag without required value
				if !flag.RequiresValue {
					flag.Exists = true
					foundFlags = append(foundFlags, flag)
				}
				// Add flag with value
				if flag.RequiresValue {
					// Pass next element as value for flag
					if i+1 > len(args) {
						return nil, fmt.Errorf("no value found for flag %s", flag.Name)
					}
					flag.Value = strings.Join(args[i+1:], " ")
					flag.Exists = true
					foundFlags = append(foundFlags, flag)
				}
			}
		}
	}

	if len(foundFlags) <= 0 {
		return nil, errors.New("no flags found")
	}

	return foundFlags, nil
}

// SendCommandHelp properly formats and shows the help page of a command
func (ctx *Context) SendCommandHelp(cmd *Command) {
	_, err := ctx.Session.ChannelMessageSendComplex(ctx.Msg.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Description: fmt.Sprintf("`%s%s` Command Help", Config.Prefix, cmd.Name),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "📜 | Description",
					Value: fmt.Sprintf("```css\n%s\n```", strings.Join(cmd.Desc, "\n")),
				},
				{
					Name:  "🤖 | Example",
					Value: fmt.Sprintf("```css\n%s\n```", strings.Join(cmd.Example, "\n")),
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("Aliases: %s", strings.Join(cmd.Aliases, " | ")),
				IconURL: ctx.Msg.Message.Author.AvatarURL("512x512"),
			},
			Color: rand.Intn(10000000),
		},
	})
	Warn(err)
}

// New collects user messages
// TODO: implement done channel
// TODO: implement reaction collector
func (collector *MessageCollector) New(ctx *Context) error {
	// Use timeout instead of channel
	if collector.UseTimeout {
		// Create timeout context
		c, cancel := context.WithTimeout(context.Background(), collector.Timeout)
		defer cancel()

	sel:
		select {
		case msg := <-ctx.LastMessage:
			// Cancel
			if msg.Timestamp >= ctx.Msg.Timestamp && msg.Author.ID == ctx.Msg.Author.ID && strings.ToLower(msg.Content) == "c" {
				return errors.New("collector canceled")
			}

			if msg.Timestamp >= ctx.Msg.Timestamp && collector.Filter(ctx, msg) {
				if collector.EndAfterOne {
					collector.MessagesCollected = append(collector.MessagesCollected, msg)
					return nil
				}
				collector.MessagesCollected = append(collector.MessagesCollected, msg)
				goto sel
			} else {
				goto sel
			}
		case <-c.Done():
			if !collector.EndAfterOne {
				return nil
			}
			return c.Err()
		}
	}

	return nil
}

// Unique remove duplicates
func Unique(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueDBUser remove duplicates
func UniqueDBUser(s []*DBUser) []*DBUser {
	keys := make(map[*DBUser]bool)
	list := []*DBUser{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueCmds remove duplicates
func UniqueCmds(m map[string]*Command) map[string]*Command {
	n := make(map[string]*Command, len(m))
	ref := make(map[*Command]bool, len(m))
	for k, v := range m {
		if _, ok := ref[v]; !ok {
			ref[v] = true
			n[k] = v
		}
	}

	return n
}

// Contains checks if slice contains a string
func Contains(s []string, e string) (bool, string) {
	for _, a := range s {
		if a == e {
			return true, a
		}
	}
	return false, ""
}

// NewDB opens the database
func NewDB() *Database {
	// Make sure file exists, if not create it
	if _, err := os.Stat("./users.yml"); err != nil {
		_ = ioutil.WriteFile("./users.yml", []byte{}, 0777)
	}
	// Open the database for reading
	db := &Database{}
	d, _ := ioutil.ReadFile("./users.yml")
	err := yaml.Unmarshal(d, &db)
	Die(err)

	return db
}

// Write writes to the database
func (db *Database) Write() {
	d, err := yaml.Marshal(&db)
	Die(err)
	err = ioutil.WriteFile("./users.yml", d, 0777)
	Die(err)
}

// AddUserToDatabase adds a user to the database
func AddUserToDatabase(dbUser *DBUser) {
	db := *NewDB()

	db[dbUser.UserID] = dbUser
	db.Write()
	return
}

// RemoveUserFromDatabase removes a user from the database
func RemoveUserFromDatabase(id string) error {
	db := *NewDB()

	if _, ok := db[id]; !ok {
		return errors.New("id is not in the database")
	}

	delete(db, id)
	db.Write()

	return nil
}
