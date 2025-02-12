package main

import (
	"flag"
	"github.com/Chrisser1/Discord-Bot-DTU/internal/utils"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Chrisser1/Discord-Bot-DTU/internal/config"
	"github.com/Chrisser1/Discord-Bot-DTU/internal/discord"
	"github.com/joho/godotenv"
)

var RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdown or not")

func init() {
	flag.Parse()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading .env file")
	} else {
		log.Println(".env file loaded successfully")
	}
	// Load the configuration
	config.LoadConfig()

	// Create pagination manager with a chosen TTL, e.g. 5 minutes.
	paginationManager := utils.NewPaginatedSessions(5 * time.Minute)

	// Initialize Discord service
	discordSvc := discord.New(config.GlobalConfig.BotToken, paginationManager)

	// Add command handlers
	discordSvc.AddCommandHandlers()

	// Start the bot
	if err := discordSvc.Start(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Register commands
	registeredCommands := discordSvc.RegisterCommands()

	// Wait for a signal to gracefully shut down the bot
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		discordSvc.RemoveCommands(registeredCommands)
	}
	// When shutting down gracefully, stop the pagination manager’s GC loop
	paginationManager.Stop()
	log.Println("Gracefully shutting down.")
}
