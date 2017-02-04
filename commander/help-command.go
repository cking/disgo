package commander

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	reListTreeSymbol = regexp.MustCompile(`├([^├]+)$`)
)

func registerHelpCommand(cmds *Command) {
	help := cmds.New("help")
	help.Description = "Search for help for a specific command"
	help.Usage = "help [command]"

	help.Handler = helpCommand
}

func helpCommand(ctx *CommandContext, res chan *CommandResponse) {
	defer close(res)

	path := strings.Fields(ctx.Content)
	prettyPath := strings.Join(path, " ")
	cmd := ctx.commander.Commands

	for _, entry := range path {
		if newCmd, ok := cmd.subCommands[entry]; ok {
			cmd = newCmd
		} else {
			res <- NewCommandErrorResponse(nil, "Command not found...")
			return
		}
	}

	rendered := cmd.Description
	if cmd.Handler != nil {
		rendered = fmt.Sprintf("`%v %v`\n", prettyPath, cmd.Usage) + rendered
	}
	rendered = rendered + renderSubcommandHelp(prettyPath, cmd)
	res <- NewCommandResponse(rendered)
}

func renderSubcommandHelp(path string, cmd *Command) string {
	if len(cmd.subCommands) == 0 {
		return ""
	}

	rendered := "\n\n**Available nested Commands**\n*(call help and the subcommand for details)*"

	entries := make([]string, 0, len(cmd.subCommands))

	for entry := range cmd.subCommands {
		entries = append(entries, entry)
	}
	sort.Strings(entries)
	for _, entry := range entries {
		subCmd := cmd.subCommands[entry]
		rendered = rendered + fmt.Sprintf("\n├ `%v %v`: %v", path, entry, strings.Split(subCmd.Description, "\n")[0])
	}

	pos := strings.LastIndex(rendered, "├")
	if pos > 0 {
		return rendered[0:pos] + "└" + rendered[pos+utf8.RuneLen('└'):]
	}
	return rendered
}
