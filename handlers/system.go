package handlers

import (
    "strings"
    "time"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/types/events"

    "whatsappbot/config"
)

// commandRegistry stores all registered commands using the Command interface
var commandRegistry = make(map[string]Command)

// RegisterCommand adds a command (that implements the Command interface) to the registry
func RegisterCommand(cmd Command) {
    // Register the main name
    commandRegistry[strings.ToLower(cmd.Name())] = cmd
    // Register aliases
    for _, alias := range cmd.Aliases() {
    commandRegistry[strings.ToLower(alias)] = cmd
    }
}

// GetCommand retrieves a command from the registry by name or alias
func GetCommand(name string) (Command, bool) {
    cmd, found := commandRegistry[strings.ToLower(name)]
    return cmd, found
}

// GetAllCommands returns a slice of all registered commands (useful for help menus)
// Returns unique commands, avoiding duplicates from aliases.
func GetAllCommands() []Command {
    // Use a map to track unique commands by their primary name
    uniqueCommands := make(map[string]Command)
    for _, cmd := range commandRegistry {
    uniqueCommands[strings.ToLower(cmd.Name())] = cmd
    }
    // Convert map values to a slice
    commandsSlice := make([]Command, 0, len(uniqueCommands))
    for _, cmd := range uniqueCommands {
    commandsSlice = append(commandsSlice, cmd)
    }
    // Optionally sort the slice by name here if needed
    return commandsSlice
}

// ProcessCommand parses and executes a command if it matches the prefix and is registered
func ProcessCommand(client *whatsmeow.Client, msg *events.Message, cfg config.Config, startTime time.Time) {
    msgText := strings.TrimSpace(msg.Message.GetConversation())

    // Check if the message starts with the command prefix
     if !strings.HasPrefix(msgText, cfg.CommandPrefix) {
     return // Not a command
     }

    // Remove prefix and split into command name and arguments
     parts := strings.Fields(msgText[len(cfg.CommandPrefix):])
     if len(parts) == 0 {
     return // Empty command after prefix
     }

     commandName := strings.ToLower(parts[0])
     args := parts[1:]

    // Look up the command in the registry
     cmd, found := GetCommand(commandName)
     if !found {
     // Optional: Send a "command not found" message
     // client.Log.Infof("Command '%s' not found", commandName)
     return
     }

    // Execute the command handler
     client.Log.Infof("Executing command '%s' for %s", commandName, msg.Info.Sender.String())
     err := cmd.Execute(client, msg, args, cfg, startTime)
     if err != nil {
     // Log the error returned by the command's Execute method
     client.Log.Errorf("Error executing command '%s': %v", commandName, err)
     // Optionally, send a generic error message to the user
     // client.SendMessage(context.Background(), msg.Info.Chat, &waProto.Message{Conversation: proto.String("Maaf, terjadi kesalahan saat menjalankan perintah.")})
     }
}

