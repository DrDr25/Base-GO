package handlers

import (
    "fmt"
    "time"

    "github.com/fatih/color"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    // waLog "go.mau.fi/whatsmeow/util/log" // Removed waLog import
)

// ANSI color codes
var (
     blue   = color.New(color.FgBlue).SprintFunc()
     yellow = color.New(color.FgYellow).SprintFunc()
     green  = color.New(color.FgGreen).SprintFunc()
     red    = color.New(color.FgRed).SprintFunc()
)

// PrintMessageLog logs incoming message details to the console with colors and message type (Reverted to fmt.Println)
func PrintMessageLog(msg *events.Message) {
     timestamp := msg.Info.Timestamp.In(time.FixedZone("WIB", 7*60*60)).Format("2006-01-02 15:04:05") // Format time to WIB
     senderName := msg.Info.PushName
     senderJID := msg.Info.Sender.String()
     groupJID := ""
     if msg.Info.Chat.Server == types.GroupServer {
     // Fetch group info if needed (might require additional client methods or storing group info)
     // For now, just use the JID
     groupJID = msg.Info.Chat.String()
     }

     messageType := "Unknown"
     messageContent := ""
     caption := ""

     if msg.Message.GetConversation() != "" {
     messageType = "Text"
     messageContent = msg.Message.GetConversation()
     } else if img := msg.Message.GetImageMessage(); img != nil {
     messageType = "Image"
     messageContent = "[Image]"
     if img.Caption != nil {
         caption = *img.Caption
     }
     } else if vid := msg.Message.GetVideoMessage(); vid != nil {
     messageType = "Video"
     messageContent = "[Video]"
     if vid.Caption != nil {
         caption = *vid.Caption
     }
     } else if sticker := msg.Message.GetStickerMessage(); sticker != nil {
     messageType = "Sticker"
     messageContent = "[Sticker]"
     } else if doc := msg.Message.GetDocumentMessage(); doc != nil {
     messageType = "Document"
     messageContent = fmt.Sprintf("[Document: %s]", doc.GetFileName())
     } else if audio := msg.Message.GetAudioMessage(); audio != nil {
     messageType = "Audio"
     messageContent = "[Audio]"
     } else if contact := msg.Message.GetContactMessage(); contact != nil {
     messageType = "Contact"
     messageContent = fmt.Sprintf("[Contact: %s]", contact.GetDisplayName())
     } else if loc := msg.Message.GetLocationMessage(); loc != nil {
     messageType = "Location"
     messageContent = "[Location]"
     } else if liveLoc := msg.Message.GetLiveLocationMessage(); liveLoc != nil {
     messageType = "LiveLocation"
     messageContent = "[Live Location]"
     } else if extText := msg.Message.GetExtendedTextMessage(); extText != nil {
     // This often wraps other messages, especially when quoted
     messageType = "ExtendedText"
     messageContent = extText.GetText()
     // You might want to recursively check the ContextInfo for the original message type
     } else {
     // Add more types as needed (Polls, Buttons, Lists, etc.)
     messageType = "Other"
     messageContent = "[Unsupported Type]"
     }

     fmt.Println(blue("\n══════ 📩 Pesan Baru 📩 ══════")) // Added newline before header
     fmt.Printf("[%s] %s (%s)\n", timestamp, yellow(senderName), green(senderJID))
     if groupJID != "" {
     fmt.Printf("  ↳ Grup 📢: %s\n", red(groupJID))
     }
     fmt.Printf("  ↳ Tipe: %s\n", messageType)
     fmt.Printf("  ↳ Pesan: %s\n", messageContent)
     if caption != "" {
     fmt.Printf("  ↳ Caption: %s\n", caption)
     }
     fmt.Println("══════════════════════════════") // Added footer
}

