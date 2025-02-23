package autocompletions

import (
	"log"
	"strings"

	"github.com/Chrisser1/Discord-Bot-DTU/internal/model"

	"github.com/bwmarrin/discordgo"
)

func CourseAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Sanity check: ensure we have the right command & option
	data := i.ApplicationCommandData()
	if data.Name != "fetch_course" {
		return
	}
	if len(data.Options) == 0 {
		return
	}

	// Get user input
	userInput := strings.ToUpper(data.Options[0].StringValue())

	// Get saved course numbers
	courseNumbers, err := model.GetSavedCourses()
	if err != nil {
		log.Println("Error fetching saved courses: ", err)
		return
	}

	// Filter matching courses
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, course := range courseNumbers {
		if strings.HasPrefix(course, userInput) {
			// Get the course number by it self
			courseNumber := strings.Split(course, ",")[0]

			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  course,
				Value: courseNumber,
			})
		}
		if len(choices) >= 25 { // Discord limits autocomplete results to 25
			break
		}
	}

	// Respond with autocomplete choices
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Println("Error sending autocomplete response:", err)
	}
}
