package commander

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cking/disgo/dge"

	"github.com/bwmarrin/discordgo"
	"github.com/cking/argparse"
)

// Command The command definition
type Command struct {
	Description string
	Usage       string

	Handler func(context *CommandContext, response chan *CommandResponse)

	subCommands map[string]*Command
}

// New Create a new Command
func (c *Command) New(name string) *Command {
	if c.subCommands == nil {
		c.subCommands = make(map[string]*Command)
	}

	newCommand := &Command{}
	c.subCommands[name] = newCommand
	return newCommand
}

// Call Execute the command
func (c *Command) Call(s *discordgo.Session, ctx *CommandContext) error {
	if c.Handler == nil {
		return errors.New("No command handler defined")
	}

	responseChannel := make(chan *CommandResponse)
	go c.Handler(ctx, responseChannel)
	for {
		response, more := <-responseChannel
		if !more {
			return nil
		}

		if response.embed != nil {
			s.ChannelMessageSendEmbed(ctx.Channel.ID, response.embed)
		} else if response.file != nil {
			if len(response.message) > 0 {
				s.ChannelFileSendWithMessage(ctx.Channel.ID, response.message, response.filename, response.file)
			} else {
				s.ChannelFileSend(ctx.Channel.ID, response.filename, response.file)
			}
		} else {
			s.ChannelMessageSend(ctx.Channel.ID, response.message)
		}
	}
}

// SetHandler sets a command handler function which parses the message content
// using the given command format
func (c *Command) SetHandler(format string, impl func(*CommandContext, chan *CommandResponse), parameters argparse.ParameterMap) *argparse.Parser {
	c.Usage = format
	cmd := argparse.NewWithoutWhitespace(format)
	cmd.SetParameters(parameters)
	cmd.SetParameter("channel", channelParameter)
	c.Handler = func(cc *CommandContext, cr chan *CommandResponse) {
		match, err := cmd.Parse(cc.Content)
		if err != nil {
			cr <- NewCommandErrorResponse(err, "failed to parse command, make sure to use the expected format of `"+c.Usage+"`")
			close(cr)
			return
		}

		// closing the channel is objective of the handler implementation
		// so no `defer` or explicit close
		cc.Params = match
		impl(cc, cr)
	}

	return cmd
}

// CommandContext The Context for command execution
type CommandContext struct {
	IsPrivate bool
	Message   *discordgo.Message
	Channel   *dge.Channel
	Guild     *dge.Guild
	Author    *dge.User
	Member    *dge.Member
	Content   string
	Params    *argparse.Match
	Session   *discordgo.Session

	commander *Commander
}

// Emoji Convert a human readable emoji code to the internal ID, or use an alternative if not found
func (cc *CommandContext) Emoji(code string, alternative string) string {
	code = strings.ToLower(code)
	if !cc.IsPrivate {
		emojis := cc.Guild.Emojis

		for _, emoji := range emojis {
			if strings.ToLower(emoji.Name) == code {
				return "<:" + code + ":" + emoji.ID + ">"
			}
		}
	}

	return alternative
}

// CommandResponse Response object for commands
type CommandResponse struct {
	message  string
	embed    *discordgo.MessageEmbed
	file     io.Reader
	filename string
}

// NewCommandResponse Create a CommandResponse with only a message text
func NewCommandResponse(text string) *CommandResponse {
	return &CommandResponse{message: text}
}

// NewCommandEmbedResponse Create a CommandResponse with only an custom embed object
func NewCommandEmbedResponse(embed *discordgo.MessageEmbed) *CommandResponse {
	return &CommandResponse{embed: embed}
}

// NewCommandErrorResponse Create a CommandResponse with an error
func NewCommandErrorResponse(err error, text string) *CommandResponse {
	if err == nil {
		return &CommandResponse{message: text}
	}
	return &CommandResponse{message: fmt.Sprintf("%v\n```%v```", text, err)}
}

// NewCommandFileResponse Create a CommandResponse with a file
func NewCommandFileResponse(file io.Reader, filename string) *CommandResponse {
	return &CommandResponse{file: file, filename: filename}
}

// NewCommandFileAndMessageResponse Create a CommandResponse with a file and a message
func NewCommandFileAndMessageResponse(file io.Reader, filename string, message string) *CommandResponse {
	return &CommandResponse{file: file, filename: filename, message: message}
}
