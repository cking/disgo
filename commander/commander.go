package commander

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/zdt/red-panda/dge"

	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
)

var (
	reWord = regexp.MustCompile(`^(\S+)`)
)

// New creates a new Commander instance
func New() *Commander {
	return &Commander{
		connected: false,
		Commands:  &Command{},
	}
}

// Commander interface
type Commander struct {
	connected     bool
	me            *discordgo.User
	mention       *regexp.Regexp
	commandChecks []func(*CommandContext) (bool, string)

	Commands *Command
}

// Connect discord with Commander
func (cmder *Commander) Connect(session *discordgo.Session) error {
	if cmder.connected {
		return fmt.Errorf("commander already connected to discord")
	}

	cmder.connected = true
	cmder.me, _ = session.User("@me")
	cmder.mention = regexp.MustCompile("^<@!?" + cmder.me.ID + ">")
	session.AddHandler(cmder.onMessageCreate)

	return nil
}

// AddCommandCheck Adds a check if a message is a command
func (cmder *Commander) AddCommandCheck(checker func(*CommandContext) (bool, string)) error {
	if cmder.connected {
		return fmt.Errorf("commander already connected to discord")
	}

	if cmder.commandChecks == nil {
		cmder.commandChecks = make([]func(*CommandContext) (bool, string), 0)
	}

	cmder.commandChecks = append(cmder.commandChecks, checker)
	return nil
}

func createCommandContext(s *discordgo.Session, m *discordgo.Message) *CommandContext {
	var err error

	ctx := &CommandContext{
		IsPrivate: true,
		Message:   m,
		Author:    &dge.User{User: m.Author},
		Content:   m.Content,
	}

	ctx.Channel, err = dge.GetChannel(s, m.ChannelID)
	if err != nil {
		panic(err)
	}

	if !ctx.Channel.IsPrivate {
		ctx.IsPrivate = false
		ctx.Guild, err = dge.GetGuild(s, ctx.Channel.GuildID)
		if err != nil {
			panic(err)
		}

		ctx.Member, err = dge.GetGuildMember(s, ctx.Guild.ID, ctx.Author.ID)
		if err != nil {
			panic(err)
		}
	}

	return ctx
}

func onMessageCreateRecover() {
	if r := recover(); r != nil {
		log.Errorf("Recovered in %v", r)
	}
}

func (cmder *Commander) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer onMessageCreateRecover()

	// ignore myself
	if m.Author.ID == cmder.me.ID {
		return
	}

	ctx := createCommandContext(s, m.Message)

	// valid command
	if cmder.mention.MatchString(ctx.Content) { // mentioned (hard coded)
		ctx.Content = strings.TrimSpace(cmder.mention.ReplaceAllString(m.Content, ""))
	} else if cap(cmder.commandChecks) > 0 { // one of the checker functions returns a valid statement
		for _, checker := range cmder.commandChecks {
			valid, messageContent := checker(ctx)
			if valid {
				ctx.Content = messageContent
				break
			}
		}
	} else {
		return
	}

	log.WithField("message", ctx.Content).Debug("Possible command incoming")
	command := cmder.Commands
	commandPath := ""
	for reWord.MatchString(ctx.Content) {
		word := reWord.FindString(ctx.Content)
		if _, ok := command.subCommands[word]; ok {
			command = command.subCommands[word]
			commandPath = commandPath + word + " "
			ctx.Content = strings.TrimSpace(reWord.ReplaceAllString(ctx.Content, ""))
		} else {
			break
		}
	}

	log.WithFields(log.Fields{
		"command": commandPath,
		"author":  ctx.Author,
		"channel": ctx.Channel,
		"guild":   ctx.Guild,
	}).Debug("Found a valid command!")
	start := time.Now()
	s.ChannelTyping(ctx.Channel.ID)
	command.Call(s, ctx)
	duration := time.Since(start)
	log.WithField("duration", duration.String()).Debug("Command executed!")
}