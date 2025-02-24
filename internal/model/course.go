package model

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Chrisser1/Discord-Bot-DTU/internal/utils"
)

// CourseScheduleSection contains schedule-related details.
type CourseScheduleSection struct {
	Schedule         string
	Location         string
	ScopeAndForm     string
	DurationOfCourse string
}

func (s CourseScheduleSection) GetSectionName() string {
	return "Schedule & Location"
}

func (s CourseScheduleSection) GetSectionValue() string {
	var sb strings.Builder
	sb.WriteString(utils.WriteLine("Schedule", s.Schedule))
	sb.WriteString(utils.WriteLine("Location", s.Location))
	sb.WriteString(utils.WriteLine("Scope & Form", s.ScopeAndForm))
	sb.WriteString(utils.WriteLine("Duration", s.DurationOfCourse))
	return sb.String()
}

func (s CourseScheduleSection) GetSectionInline() bool {
	return false
}

// SetInLine implements utils.Section.
func (s CourseScheduleSection) SetInLine(IsInLine bool) {
	panic("CourseScheduleSection does not support inline formatting")
}

// CourseExamSection contains examination-related details.
type CourseExamSection struct {
	DateOfExamination string
	TypeOfAssessment  string
	ExamDuration      string
	Aid               string
	Evaluation        string
}

func (s CourseExamSection) GetSectionName() string {
	return "Examination Details"
}

func (s CourseExamSection) GetSectionValue() string {
	var sb strings.Builder
	sb.WriteString(utils.WriteLine("Date of Examination", s.DateOfExamination))
	sb.WriteString(utils.WriteLine("Type of Assessment", s.TypeOfAssessment))
	sb.WriteString(utils.WriteLine("Exam Duration", s.ExamDuration))
	sb.WriteString(utils.WriteLine("Aid", s.Aid))
	sb.WriteString(utils.WriteLine("Evaluation", s.Evaluation))
	return sb.String()
}

func (s CourseExamSection) GetSectionInline() bool {
	return false
}

// SetInLine implements utils.Section.
func (s CourseExamSection) SetInLine(IsInLine bool) {
	panic("CourseExamSection does not support inline formatting")
}

// CourseResponsibleSection contains the responsible and co-responsible teachers.
type CourseResponsibleSection struct {
	Responsible         string
	CourseCoResponsible string
}

func (s CourseResponsibleSection) GetSectionName() string {
	return "Responsible Teachers"
}

func (s CourseResponsibleSection) GetSectionValue() string {
	var sb strings.Builder
	sb.WriteString(utils.WriteLine("Responsible", s.Responsible))
	sb.WriteString(utils.WriteLine("Co-Responsible", s.CourseCoResponsible))
	return sb.String()
}

func (s CourseResponsibleSection) GetSectionInline() bool {
	return false
}

// SetInLine implements utils.Section.
func (s CourseResponsibleSection) SetInLine(IsInLine bool) {
	panic("CourseResponsibleSection does not support inline formatting")
}

// CourseAdditionalSection contains all the remaining course details.
type CourseAdditionalSection struct {
	NotApplicableTogetherWith string
	AcademicPrerequisites     string
	Department                string
	DepartmentInvolved        string
	HomePage                  string
	RegistrationSignUp        string
	FetchTime                 time.Time
}

func (s CourseAdditionalSection) GetSectionName() string {
	return "Additional Information"
}

func (s CourseAdditionalSection) GetSectionValue() string {
	var sb strings.Builder
	sb.WriteString(utils.WriteLine("Not Applicable Together With", s.NotApplicableTogetherWith))
	sb.WriteString(utils.WriteLine("Academic Prerequisites", s.AcademicPrerequisites))
	sb.WriteString(utils.WriteLine("Department", s.Department))
	sb.WriteString(utils.WriteLine("Department Involved", s.DepartmentInvolved))
	sb.WriteString(utils.WriteLine("Home Page", s.HomePage))
	sb.WriteString(utils.WriteLine("Registration Sign-Up", s.RegistrationSignUp))
	sb.WriteString(utils.WriteLine("Fetched", s.FetchTime.Format(time.RFC1123)))
	return sb.String()
}

func (s CourseAdditionalSection) GetSectionInline() bool {
	return false
}

// SetInLine implements utils.Section.
func (s CourseAdditionalSection) SetInLine(IsInLine bool) {
	panic("CourseAdditionalSection does not support inline formatting")
}

type CourseTypeSection struct {
	CourseType []utils.CourseTypeBlock
}

// Course represents the course details.
type Course struct {
	CourseNumber             string
	Title                    string
	DanishTitle              string
	LanguageOfInstruction    string
	ECTS                     string
	CourseTypeSection        CourseTypeSection
	CourseScheduleSection    CourseScheduleSection
	CourseExamSection        CourseExamSection
	CourseResponsibleSection CourseResponsibleSection
	CourseAdditionalSection  CourseAdditionalSection
	IsInLine                 bool // New field for inline formatting
}

// GetSectionName returns a formatted section header for Discord
func (c *Course) GetSectionName() string {
	return fmt.Sprintf("**Course: %s - %s**", c.CourseNumber, c.Title)
}

// GetSectionValue returns a formatted course description
func (c *Course) GetSectionValue() string {
	var sb strings.Builder

	// Always show the Title as a header.
	sb.WriteString(utils.WriteLine("Danish Title", c.DanishTitle))
	sb.WriteString(utils.WriteLine("Language", c.LanguageOfInstruction))
	sb.WriteString(utils.WriteLine("ECTS", c.ECTS))

	return sb.String()
}

// GetSectionInline returns whether the section should be inline
func (c *Course) GetSectionInline() bool {
	return c.IsInLine
}

// SetInLine sets the inline formatting for Discord embeds
func (c *Course) SetInLine(isInLine bool) {
	c.IsInLine = isInLine
}

// FetchCourse retrieves the *rendered* DTU course page and parses it into a Course struct.
func FetchCourse(courseNumber string) (*Course, error) {
	// Build the course URL
	url := fmt.Sprintf("https://kurser.dtu.dk/course/%s", courseNumber)

	// Use the dynamic fetcher from utils
	doc, err := utils.FetchDynamicCoursePage(url)
	if err != nil {
		return nil, err
	}

	// Check if the main info is actually present
	if doc.Find("div.box.information").Length() == 0 {
		log.Println("No course information found on the rendered page!")
		return nil, nil
	}

	// Create a new Course object
	course := &Course{
		CourseNumber: courseNumber,
		CourseAdditionalSection: CourseAdditionalSection{
			FetchTime: time.Now(),
		},
	}

	// Parse the course details
	rawTitle := utils.ExtractText(doc, "div.col-xs-8 h2", "")
	// We remove the leading course number from the title e.g., "10060 Physics (Polytechnical Foundation) -> Physics (Polytechnical Foundation)"
	course.Title = strings.TrimSpace(strings.SplitN(rawTitle, " ", 2)[1])
	course.DanishTitle = utils.ExtractFieldByName(doc, "Danish title")
	course.LanguageOfInstruction = utils.ExtractFieldByName(doc, "Language of instruction")
	course.ECTS = utils.ExtractFieldByName(doc, "Point( ECTS )")
	course.CourseTypeSection.CourseType = utils.ExtractCourseTypeAdvanced(doc)
	course.CourseScheduleSection.Schedule = utils.ExtractFieldByName(doc, "Schedule")
	course.CourseScheduleSection.Location = utils.ExtractFieldByName(doc, "Location")
	course.CourseScheduleSection.ScopeAndForm = utils.ExtractFieldByName(doc, "Scope and form")
	course.CourseScheduleSection.DurationOfCourse = utils.ExtractFieldByName(doc, "Duration of course")
	course.CourseExamSection.DateOfExamination = utils.ExtractFieldByName(doc, "Date of examination")
	course.CourseExamSection.TypeOfAssessment = utils.ExtractFieldByName(doc, "Type of assessment")
	course.CourseExamSection.ExamDuration = utils.ExtractFieldByName(doc, "Exam duration")
	course.CourseExamSection.Aid = utils.ExtractFieldByName(doc, "Aid")
	course.CourseExamSection.Evaluation = utils.ExtractFieldByName(doc, "Evaluation")
	course.CourseAdditionalSection.NotApplicableTogetherWith = utils.ExtractFieldByName(doc, "Not applicable together with")
	course.CourseAdditionalSection.AcademicPrerequisites = utils.ExtractFieldByName(doc, "Academic prerequisites")
	course.CourseResponsibleSection.Responsible = utils.ExtractFieldByName(doc, "Responsible")
	course.CourseResponsibleSection.CourseCoResponsible = utils.ExtractFieldByName(doc, "Course co-responsible")
	course.CourseAdditionalSection.Department = utils.ExtractFieldByName(doc, "Department")
	course.CourseAdditionalSection.DepartmentInvolved = utils.ExtractFieldByName(doc, "Department involved")
	course.CourseAdditionalSection.HomePage = utils.ExtractFieldByName(doc, "Home page")
	course.CourseAdditionalSection.RegistrationSignUp = utils.ExtractFieldByName(doc, "Registration sign-up")

	// Validate that we actually found some data
	if course.Title == "" && course.DanishTitle == "" && course.ECTS == "" {
		log.Println("No valid data received for course:", courseNumber)
		return nil, nil
	}

	return course, nil
}

// SaveCourseToFile appends the given course string (e.g., "01001, Mathematics 1a") to "data/courses.txt"
func SaveCourseToFile(course string) error {
	// Read the existing file (if it exists) to check for duplicates.
	data, err := os.ReadFile("data/courses.txt")
	if err == nil {
		// Split the file content into lines and check each one.
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == course {
				// Course already exists; nothing to do.
				return nil
			}
		}
	} else if !os.IsNotExist(err) {
		// An error other than file not existing occurred.
		return err
	}

	// Open file in append mode; create it if it doesn't exist.
	f, err := os.OpenFile("data/courses.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the course string followed by a newline.
	if _, err = f.WriteString(course + "\n"); err != nil {
		return err
	}

	return nil
}

// GetSavedCourses reads "data/courses.txt" and returns a slice of non-empty course lines.
func GetSavedCourses() ([]string, error) {
	data, err := os.ReadFile("data/courses.txt")
	if err != nil {
		return nil, err
	}

	// Split the file content by newline.
	lines := strings.Split(string(data), "\n")
	var courses []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			courses = append(courses, trimmed)
		}
	}
	return courses, nil
}
