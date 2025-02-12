package model

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Course represents the course details.
type Course struct {
	CourseNumber      string `json:"course_number"`
	Title             string `json:"title"`
	EnglishTitle      string `json:"english_title"`
	Language          string `json:"language"`
	ECTS              string `json:"ects"`
	Type              string `json:"type"`
	Schedule          string `json:"schedule"`
	Location          string `json:"location"`
	TeachingFormat    string `json:"teaching_format"`
	Duration          string `json:"duration"`
	ExamPlacement     string `json:"exam_placement"`
	Evaluation        string `json:"evaluation"`
	ExamDuration      string `json:"exam_duration"`
	AllowedMaterials  string `json:"allowed_materials"`
	GradingScale      string `json:"grading_scale"`
	PointReference    string `json:"point_reference"`
	Prerequisites     string `json:"prerequisites"`
	CourseResponsible string `json:"course_responsible"`
	CoResponsible     string `json:"co_responsible"`
	Institute         string `json:"institute"`
	Website           string `json:"website"`
	IsInLine          bool   `json:"is_in_line"` // New field for inline formatting
	FetchTime         time.Time
}

// GetSectionName returns a formatted section header for Discord
func (c *Course) GetSectionName() string {
	return fmt.Sprintf("**Course: %s**", c.CourseNumber)
}

// GetSectionValue returns a formatted course description
func (c *Course) GetSectionValue() string {
	return fmt.Sprintf(
		"> **Title**: %s\n"+
			"> **ECTS**: `%s`\n"+
			"> **Language**: `%s`\n"+
			"> **Course Type**: `%s`\n"+
			"> **Schedule**: `%s`\n"+
			"> **Location**: `%s`\n"+
			"> **Teaching Format**: `%s`\n"+
			"> **Duration**: `%s`\n"+
			"> **Exam Placement**: `%s`\n"+
			"> **Evaluation**: `%s`\n"+
			"> **Exam Duration**: `%s`\n"+
			"> **Allowed Materials**: `%s`\n"+
			"> **Grading Scale**: `%s`\n"+
			"> **Point Reference**: `%s`\n"+
			"> **Prerequisites**: `%s`\n"+
			"> **Course Responsible**: `%s`\n"+
			"> **Co-Responsible**: `%s`\n"+
			"> **Institute**: `%s`\n"+
			"> **Fetched**: `%s`\n"+
			"> **Website**: [Course Page](%s)",
		c.Title, c.ECTS, c.Language, c.Type, c.Schedule, c.Location, c.TeachingFormat,
		c.Duration, c.ExamPlacement, c.Evaluation, c.ExamDuration, c.AllowedMaterials,
		c.GradingScale, c.PointReference, c.Prerequisites, c.CourseResponsible,
		c.CoResponsible, c.Institute, c.FetchTime.Format(time.RFC1123), c.Website,
	)
}

// GetSectionInline returns whether the section should be inline
func (c *Course) GetSectionInline() bool {
	return c.IsInLine
}

// SetInLine sets the inline formatting for Discord embeds
func (c *Course) SetInLine(isInLine bool) {
	c.IsInLine = isInLine
}

// FetchCourse retrieves course data from DTU's course website.
func FetchCourse(courseNumber string) (*Course, error) {
	url := fmt.Sprintf("https://kurser.dtu.dk/course/%s", courseNumber)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch course: %s (status: %d)", courseNumber, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Debugging: Check if we are actually fetching a course page
	pageTitle := doc.Find("title").Text()
	log.Println("Fetched Page Title:", pageTitle)

	// Ensure we are on a valid course page
	if doc.Find("div.box.information").Length() == 0 {
		log.Println("No course information found on the page!")
		return nil, nil
	}

	course := &Course{CourseNumber: courseNumber}

	// Extract course details using more precise selectors
	course.Title = extractText(doc, "div.box.information h1")
	course.EnglishTitle = extractText(doc, "td:has(label:contains('Engelsk titel')) + td")
	course.Language = extractText(doc, "td:has(label:contains('Undervisningssprog')) + td")
	course.ECTS = extractText(doc, "td:has(label:contains('Point')) + td")

	// Extract Kursustype, including hidden elements
	course.Type = extractCourseType(doc)

	course.Schedule = extractText(doc, "td:has(label a:contains('Skemaplacering')) + td")
	course.Location = extractText(doc, "td:has(label:contains('Undervisningens placering')) + td")
	course.TeachingFormat = extractText(doc, "td:has(label:contains('Undervisningsform')) + td")
	course.Duration = extractText(doc, "td:has(label:contains('Kursets varighed')) + td")
	course.ExamPlacement = extractText(doc, "td:has(label a:contains('Eksamensplacering')) + td")
	course.Evaluation = extractText(doc, "td:has(label:contains('Evalueringsform')) + td")
	course.ExamDuration = extractText(doc, "td:has(label:contains('Eksamensvarighed')) + td")
	course.AllowedMaterials = extractText(doc, "td:has(label:contains('Hjælpemidler')) + td")
	course.GradingScale = extractText(doc, "td:has(label:contains('Bedømmelsesform')) + td")
	course.PointReference = extractText(doc, "td:has(label:contains('Pointspærring')) + td")

	// Extract prerequisites (handling multiple links inside <td>)
	course.Prerequisites = extractPrerequisites(doc)

	course.CourseResponsible = extractText(doc, "td:has(label:contains('Kursusansvarlig')) + td")
	course.CoResponsible = extractText(doc, "td:has(label:contains('Medansvarlige')) + td")
	course.Institute = extractText(doc, "td:has(label:contains('Institut')) + td")
	course.Website = extractCourseWebsite(doc)

	// Ensure we only return valid courses (not empty results)
	if course.Title == "" || course.EnglishTitle == "" || course.ECTS == "" {
		log.Println("No valid data received for course:", courseNumber)
		return nil, nil
	}

	return course, nil
}

// extractText safely gets the text of an element
func extractText(doc *goquery.Document, selector string) string {
	text := doc.Find(selector).Text()
	return strings.TrimSpace(text)
}

// extractCourseType handles extracting Kursustype, including hidden elements
func extractCourseType(doc *goquery.Document) string {
	var types []string
	doc.Find("td:has(label:contains('Kursustype')) + td div").Each(func(i int, s *goquery.Selection) {
		// Append both visible and hidden elements
		s.Children().Each(func(j int, span *goquery.Selection) {
			types = append(types, strings.TrimSpace(span.Text()))
		})
	})
	return strings.Join(types, ", ") // Join multiple course types with commas
}

// extractPrerequisites extracts multiple course prerequisite links
func extractPrerequisites(doc *goquery.Document) string {
	var prereq []string
	doc.Find("td:has(label:contains('Faglige forudsætninger')) a.CourseLink").Each(func(i int, s *goquery.Selection) {
		prereq = append(prereq, strings.TrimSpace(s.Text()))
	})
	return strings.Join(prereq, ", ")
}

// extractCourseWebsite gets the website link for the course
func extractCourseWebsite(doc *goquery.Document) string {
	link, exists := doc.Find("td:has(label:contains('Kursushjemmeside')) a").Attr("href")
	if exists {
		return strings.TrimSpace(link)
	}
	return ""
}

func SaveCourseToFile(course *Course) error {
	filename := fmt.Sprintf("courses/%s.json", course.CourseNumber)
	data, err := json.MarshalIndent(course, "", "  ")
	if err != nil {
		return err
	}

	// Ensure "courses" directory exists
	if _, err := os.Stat("courses"); os.IsNotExist(err) {
		os.Mkdir("courses", os.ModePerm)
	}

	return os.WriteFile(filename, data, 0644)
}

func GetSavedCourseNumbers() ([]string, error) {
	files, err := os.ReadDir("courses")
	if err != nil {
		return nil, err
	}

	var courseNumbers []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			courseNumbers = append(courseNumbers, strings.TrimSuffix(file.Name(), ".json"))
		}
	}
	return courseNumbers, nil
}
