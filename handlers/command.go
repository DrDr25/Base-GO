package handlers

import (
    "time"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsappbot/config"
)

// Command defines the interface that all bot commands must implement
type Command interface {
    // Execute runs the command logic
    Execute(client *whatsmeow.Client, msg *events.Message, args []string, cfg config.Config, startTime time.Time) error
    // Name returns the primary name of the command
    Name() string
    // Description returns a short description of the command
    Description() string
    // Aliases returns a list of alternative names for the command
    Aliases() []string
    // Category returns the category of the command (e.g., "Info", "Tools")
    Category() string
    // Usage returns how to use the command (optional, can return "")
    // Usage() string
}

