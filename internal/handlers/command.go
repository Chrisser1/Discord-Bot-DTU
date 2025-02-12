package handlers

import (
	"github.com/Chrisser1/Discord-Bot-DTU/internal/handlers/autocompletions"
	"github.com/Chrisser1/Discord-Bot-DTU/internal/handlers/commands"
	"github.com/Chrisser1/Discord-Bot-DTU/internal/utils"
	"github.com/bwmarrin/discordgo"
)

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "fetch_course",
			Description: "Fetches a specific dtu course",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "course_code",
					Description: "The course code to fetch",
					Required:    true,
					Autocomplete: true,
				},
			},
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, pm *utils.PaginatedSessions){
		"fetch_course":        commands.FetchCourse,
	}

	AutocompleteHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"fetch_course": autocompletions.CourseAutocomplete,
	}
)
