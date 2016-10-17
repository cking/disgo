package dge

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Channel Channel type override
type Channel struct {
	*discordgo.Channel
}

func (c *Channel) String() string {
	return fmt.Sprintf("%v <%v>", c.Name, c.ID)
}

// Guild Guild type override
type Guild struct {
	*discordgo.Guild
}

func (g *Guild) String() string {
	return fmt.Sprintf("%v <%v>", g.Name, g.ID)
}

// Member Member type override
type Member struct {
	*discordgo.Member
}

// DgeUser Convert the embedded user object to a DGE override
func (m *Member) DgeUser() *User {
	return &User{User: m.User}
}

func (m *Member) String() string {
	return fmt.Sprintf("%v (%v)", m.DgeUser(), m.Nick)
}

// User User type override
type User struct {
	*discordgo.User
}

func (u *User) String() string {
	return fmt.Sprintf("%v#%v (%v)", u.Username, u.Discriminator, u.ID)
}

// GetChannel Get a Channel from a discordgo.Session and convert it to the Dge Variant
func GetChannel(s *discordgo.Session, id string) (*Channel, error) {
	c, err := s.Channel(id)
	if err != nil {
		return nil, err
	}

	return &Channel{Channel: c}, nil
}

// GetGuild Get a Guild from a discordgo.Session and convert it to the Dge Variant
func GetGuild(s *discordgo.Session, id string) (*Guild, error) {
	c, err := s.Guild(id)
	if err != nil {
		return nil, err
	}

	return &Guild{Guild: c}, nil
}

// GetGuildMember Get a guild Member from a discordgo.Session and convert it to the Dge Variant
func GetGuildMember(s *discordgo.Session, guildID string, id string) (*Member, error) {
	c, err := s.GuildMember(guildID, id)
	if err != nil {
		return nil, err
	}

	return &Member{Member: c}, nil
}
