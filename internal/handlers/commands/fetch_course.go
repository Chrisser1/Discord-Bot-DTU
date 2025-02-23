package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/Chrisser1/Discord-Bot-DTU/internal/model"
	"github.com/Chrisser1/Discord-Bot-DTU/internal/utils"
	"github.com/bwmarrin/discordgo"
)

func FetchCourse(s *discordgo.Session, i *discordgo.InteractionCreate, pm *utils.PaginatedSessions) {
	courseID := i.ApplicationCommandData().Options[0].StringValue()

	// Fetch the course
	course, err := model.FetchCourse(courseID)
	if err != nil {
		log.Printf("Error fetching course: %v", err)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error fetching event data.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if course == nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("No course found for ID: %s", courseID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Save the course as course number, title
	err = model.SaveCourseToFile(fmt.Sprintf("%d, %s", course.CourseNumber, course.Title))
	if err != nil {
		log.Println("Failed to save course to file:", err)
	}

	// Create a new paginated session
	paginationID := utils.BuildPaginationID()

	fields := make([]utils.Section, 0)

	// log the course GetSectionValue() lengths
	fields = append(fields, course)
	fields = append(fields, course.CourseScheduleSection)
	fields = append(fields, course.CourseExamSection)
	fields = append(fields, course.CourseResponsibleSection)
	fields = append(fields, course.CourseAdditionalSection)
	for _, courseType := range course.CourseTypeSection.CourseType {
		fields = append(fields, courseType)
	}

	// Create and store the PaginationData
	data := &utils.PaginationData{
		Fields:      fields,
		PageIndex:   0,
		Description: "",
		AuthorID:    i.Member.User.ID,
		Title:       fmt.Sprintf("Fetched course: %s - %s", courseID, course.Title),
		Footer:      fmt.Sprintf("Fetched from %s", fmt.Sprintf("https://kurser.dtu.dk/course/%d", course.CourseNumber)),
		Color:       0x606060,
		CreatedAt:   time.Now(),
		PageSize:    5,
	}
	pm.Put(paginationID, data)

	if err := utils.SendInitialPaginationResponse(s, i, paginationID, data); err != nil {
		log.Println("Failed to respond with course embed:", err)
	}
}
