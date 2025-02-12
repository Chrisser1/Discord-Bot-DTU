package utils

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

// PaginationData holds pages and current state for a single pagination session.
type PaginationData struct {
	Fields      []Section
	PageSize    int
	PageIndex   int
	Description string
	AuthorID    string
	Title       string
	Footer      string
	Color       int
	CreatedAt   time.Time
	LastAccess  time.Time
}

// GetPageAmount returns how many pages we have
func (p *PaginationData) GetPageAmount() int {
	// Avoid dividing by zero if p.PageSize is 0
	if p.PageSize <= 0 {
		return 1
	}
	pages := (len(p.Fields) + p.PageSize - 1) / p.PageSize // integer ceil
	if pages < 1 {
		pages = 1
	}
	return pages
}

// PaginatedSessions manages PaginationData in a concurrent-safe way.
// It also performs periodic cleanup of stale entries based on a TTL.
type PaginatedSessions struct {
	sessions sync.Map      // key: string (paginationID), value: *PaginationData
	ttl      time.Duration // how long until we consider a session expired
	stopChan chan struct{} // channel to signal the GC goroutine to stop
}

// NewPaginatedSessions initializes a session manager with the given TTL.
func NewPaginatedSessions(ttl time.Duration) *PaginatedSessions {
	mgr := &PaginatedSessions{
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}
	// Start a background goroutine to periodically clean up stale sessions.
	go mgr.gcLoop()
	return mgr
}

// Put adds (or updates) a pagination session in the manager.
// It also sets LastAccess to "now."
func (ps *PaginatedSessions) Put(id string, data *PaginationData) {
	data.LastAccess = time.Now()
	ps.sessions.Store(id, data)
}

// Get retrieves a pagination session by ID and updates its LastAccess time.
// Returns (data, true) if found, or (nil, false) if not found/expired.
func (ps *PaginatedSessions) Get(id string) (*PaginationData, bool) {
	val, ok := ps.sessions.Load(id)
	if !ok {
		return nil, false
	}
	pd, _ := val.(*PaginationData)
	// Update last access time
	pd.LastAccess = time.Now()
	return pd, true
}

// Delete removes a session from the manager by ID.
func (ps *PaginatedSessions) Delete(id string) {
	ps.sessions.Delete(id)
}

// Stop signals the GC goroutine to stop running (e.g., on bot shutdown).
func (ps *PaginatedSessions) Stop() {
	close(ps.stopChan)
}

// gcLoop runs periodically to clean up stale sessions that exceed the TTL.
func (ps *PaginatedSessions) gcLoop() {
	// Adjust this ticker interval as needed (every minute, 30s, etc.).
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			ps.sessions.Range(func(key, value interface{}) bool {
				pd, ok := value.(*PaginationData)
				if !ok {
					// data is invalid, remove it
					ps.sessions.Delete(key)
					return true
				}
				// If last used more than ps.ttl ago, remove it
				if now.Sub(pd.LastAccess) > ps.ttl {
					log.Printf("[Pagination GC] removing stale paginationID=%s\n", key)
					ps.sessions.Delete(key)
				}
				return true
			})

		case <-ps.stopChan:
			// Manager stopped; exit the GC loop.
			return
		}
	}
}

// BuildPaginationID creates a unique ID for this pagination session
func BuildPaginationID() string {
	return uuid.NewString()
}

// MakePaginationComponents builds a row of "prev" / "next" buttons
func MakePaginationComponents(paginationID string, pageIndex, totalPages int) []discordgo.MessageComponent {
	prevDisabled := pageIndex <= 0
	nextDisabled := pageIndex >= totalPages-1

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Previous",
					Style:    discordgo.PrimaryButton,
					CustomID: "page_prev_" + paginationID,
					Disabled: prevDisabled,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: "page_next_" + paginationID,
					Disabled: nextDisabled,
				},
			},
		},
	}
}

// MakePaginationEmbed builds a simple embed for the specified page
func MakePaginationEmbed(data *PaginationData) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField

	// Fill up to PageSize fields, starting at the correct offset:
	start := data.PageIndex * data.PageSize
	for i := 0; i < data.PageSize; i++ {
		idx := start + i
		if idx >= len(data.Fields) {
			break
		}

		f := data.Fields[idx]
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   f.GetSectionName(),
			Value:  f.GetSectionValue(),
			Inline: f.GetSectionInline(),
		})
	}

	pageCount := data.GetPageAmount()
	return &discordgo.MessageEmbed{
		Title:       data.Title,
		Description: data.Description,
		Color:       data.Color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page: %d / %d", data.PageIndex+1, pageCount),
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Fields:    fields,
	}
}

// SendInitialPaginationResponse sends the first paginated message to the channel.
// It builds the embed, buttons, and uses InteractionRespond to send.
func SendInitialPaginationResponse(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	paginationID string,
	data *PaginationData,
) error {
	embed := MakePaginationEmbed(data)
	totalPages := data.GetPageAmount()
	if totalPages < 1 {
		totalPages = 1 // At least 1 page even if no fields
	}

	components := MakePaginationComponents(paginationID, data.PageIndex, totalPages)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
	if err != nil {
		log.Println("Failed to respond with paginated embed:", err)
	}

	return err
}
