package commands

import (
	"context"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
)

// IsAllowed determines if the given user has a role set out in the permissions config
func IsAllowed(ctx context.Context, s *discordgo.Session, authorID string, command string) bool {
	l, loggerOk := logging.FromContext(ctx)
	conf := config.GetConfig()
	member, err := s.GuildMember(conf.GuildID, authorID)

	if err != nil && loggerOk {
		l.Infof("Failed to retrieve Member info. Error: %q", err)
	}

	if _, hasCommand := conf.Permissions[command]; !hasCommand {
		return true
	}

	// Check the user has a role defined in the config for this command
	isAllowed := false
	for _, role := range member.Roles {
		state := s.State

		roleInfo, err := state.Role(conf.GuildID, role)
		if err != nil {
			if loggerOk {
				l.Infof("Failed to retrieve role information: \n GuildID: %q \n Role: %q \n err: %q", conf.GuildID, role, err)
			}

			isAllowed = false
			break
		}

		if StringInSlice(roleInfo.Name, conf.Permissions[command]) {
			isAllowed = true
			break
		}
	}
	return isAllowed
}

// StringInSlice searches for a given value in a flat slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
