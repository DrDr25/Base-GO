package menu

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"

    "go.mau.fi/whatsmeow"
    waProto "go.mau.fi/whatsmeow/binary/proto"
    "go.mau.fi/whatsmeow/types/events"
    "google.golang.org/protobuf/proto"

    "whatsappbot/config"  // Import config package
    "whatsappbot/handlers" // Import handlers package for registration
)

// Define a struct for the menu command
type menuCommand struct{}

// init function runs when the package is imported, used to register commands
func init() {
    // Register an instance of menuCommand
     handlers.RegisterCommand(&menuCommand{})
}

// Name returns the command name
func (cmd *menuCommand) Name() string {
    return "menu"
}

// Description returns the command description
func (cmd *menuCommand) Description() string {
    return "Menampilkan menu bot"
}

// Aliases returns command aliases
func (cmd *menuCommand) Aliases() []string {
    return []string{"help"}
}

// Category returns the command category
func (cmd *menuCommand) Category() string {
    return "Info" // Assigning menu to the Info category
}

// Execute runs the command logic
func (cmd *menuCommand) Execute(client *whatsmeow.Client, msg *events.Message, args []string, cfg config.Config, startTime time.Time) error {
    // The menu command doesn't use arguments, but they are passed in case needed in the future
     err := sendMenuWithImage(client, msg, startTime, cfg)
     if err != nil {
     // Return the error to be logged by the caller (ProcessCommand)
     return fmt.Errorf("gagal mengirim menu (Execute): %w", err)
     }
     return nil // Success
}

// Fungsi untuk mendapatkan path absolut folder tempat executable bot berada
func getExecutableDir() (string, error) {
     execPath, err := os.Executable()
     if err != nil {
     return "", fmt.Errorf("gagal mendapatkan path executable: %w", err)
     }
     return filepath.Dir(execPath), nil
}

// Fungsi untuk format durasi uptime
func formatDuration(d time.Duration) string {
     d = d.Round(time.Second)
     h := d / time.Hour
     d -= h * time.Hour
     m := d / time.Minute
     d -= m * time.Minute
     s := d / time.Second
     return fmt.Sprintf("%d jam, %d menit, %d detik", h, m, s)
}

// sendMenuWithImage sends the menu message with image and formatted caption
// startTime is the bot's start time
// cfg contains the loaded configuration
// Returns an error if sending fails (including fallback attempt)
func sendMenuWithImage(client *whatsmeow.Client, msg *events.Message, startTime time.Time, cfg config.Config) error {
    // Load timezone WIB (Asia/Jakarta)
    loc, err := time.LoadLocation("Asia/Jakarta")
     if err != nil {
     client.Log.Warnf("Gagal load timezone Asia/Jakarta, fallback ke UTC: %v", err)
     loc = time.UTC // Fallback ke UTC jika gagal
     }
     currentTime := time.Now().In(loc)

    // Dapatkan direktori executable
     execDir, err := getExecutableDir()
     if err != nil {
     client.Log.Errorf("Gagal mendapatkan direktori executable: %v", err)
     // Attempt fallback and return its error
     return sendTextMenuFallback(client, msg, loc, cfg)
     }

    // Buat path gambar relatif terhadap executable
     imagePath := filepath.Join(execDir, "image", "thumbnail.jpg")

    // Baca file gambar
     imageData, err := ioutil.ReadFile(imagePath)
     if err != nil {
     client.Log.Errorf("Gagal baca file gambar di %s: %v", imagePath, err)
     // Attempt fallback and return its error
     return sendTextMenuFallback(client, msg, loc, cfg)
     }

    // Upload gambar
     uploaded, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
     if err != nil {
     client.Log.Errorf("Gagal upload gambar: %v", err)
     // Attempt fallback and return its error
     return sendTextMenuFallback(client, msg, loc, cfg)
     }

    // Hitung uptime
     uptime := time.Since(startTime)

    // Format owner mentions (ambil dari config)
     ownerMentions := []string{}
     for _, owner := range cfg.Owners {
     ownerMentions = append(ownerMentions, "@"+owner)
     }

    // Buat caption yang lebih rapi (gunakan botName dari config)
     captionBuilder := &strings.Builder{}
     fmt.Fprintf(captionBuilder, "👋 Hai *%s*!\n\n", msg.Info.PushName)
     fmt.Fprintf(captionBuilder, "*🤖 %s BOT MENU 🤖*\n\n", strings.ToUpper(cfg.BotName))
     fmt.Fprintf(captionBuilder, "*Nama Bot*: %s\n", cfg.BotName)
     fmt.Fprintf(captionBuilder, "*Owner*: %s\n", strings.Join(ownerMentions, ", "))
     fmt.Fprintf(captionBuilder, "*Waktu Server*: %s WIB\n", currentTime.Format("15:04:05"))
     fmt.Fprintf(captionBuilder, "*Tanggal Server*: %s\n", currentTime.Format("02/01/2006"))
     fmt.Fprintf(captionBuilder, "*Uptime*: %s\n\n", formatDuration(uptime))
     // fmt.Fprintf(captionBuilder, "Berikut daftar fitur yang tersedia:\n") // Removed old header

     // Group commands by category
     allCommands := handlers.GetAllCommands()
     commandsByCategory := make(map[string][]handlers.Command)
     for _, cmd := range allCommands {
     category := cmd.Category()
     if category == "" {
         category = "Lainnya" // Default category if empty
     }
     commandsByCategory[category] = append(commandsByCategory[category], cmd)
     }

     // Get sorted category names
     categories := make([]string, 0, len(commandsByCategory))
     for category := range commandsByCategory {
     categories = append(categories, category)
     }
     // Sort categories alphabetically (optional, but nice)
     // sort.Strings(categories) // Requires importing "sort"

     // Build the categorized menu list
     for _, category := range categories {
     fmt.Fprintf(captionBuilder, "\n*╭─「 %s 」*\n", strings.ToUpper(category))
     for _, cmd := range commandsByCategory[category] {
         fmt.Fprintf(captionBuilder, "*│* `%s%s` - %s\n", cfg.CommandPrefix, cmd.Name(), cmd.Description())
     }
     fmt.Fprintf(captionBuilder, "*╰──────────*\n") // Use simple line for bottom border
     }

     caption := captionBuilder.String()

    // Prepare mentioned JIDs
     mentionedJIDs := []string{}
     for _, owner := range cfg.Owners {
     // Ensure owner JIDs are in the correct format (e.g., 628...s.whatsapp.net)
     // Assuming cfg.Owners already contains valid JIDs
     mentionedJIDs = append(mentionedJIDs, owner)
     }

    // Buat ContextInfo untuk quote dan mention
     contextInfo := &waProto.ContextInfo{
     StanzaID:      proto.String(msg.Info.ID),
     Participant:   proto.String(msg.Info.Sender.String()),
     QuotedMessage: msg.Message, // Quote pesan asli
     MentionedJID:  mentionedJIDs, // Add mentioned JIDs here
     }

    // Buat ImageMessage
     imageMessage := &waProto.ImageMessage{
     URL:           proto.String(uploaded.URL),
     Mimetype:      proto.String("image/jpeg"),
     Caption:       proto.String(caption),
     FileSHA256:    uploaded.FileSHA256,
     FileLength:    proto.Uint64(uploaded.FileLength),
     MediaKey:      uploaded.MediaKey,
     FileEncSHA256: uploaded.FileEncSHA256,
     DirectPath:    proto.String(uploaded.DirectPath),
     ContextInfo:   contextInfo,
     }

    // Kirim pesan
     _, err = client.SendMessage(context.Background(), msg.Info.Chat, &waProto.Message{
     ImageMessage: imageMessage,
     })
     if err != nil {
     // Return the error instead of just logging
     return fmt.Errorf("gagal kirim menu dengan gambar: %w", err)
     } else {
     client.Log.Infof("Menu dengan gambar terkirim ke %s!", msg.Info.Chat)
     return nil // Success
     }
}

// Fungsi fallback kalau gambar gagal diupload atau ada error lain
// Returns an error if sending the fallback fails
func sendTextMenuFallback(client *whatsmeow.Client, msg *events.Message, loc *time.Location, cfg config.Config) error {
     client.Log.Warnf("Mengirim fallback menu teks ke %s", msg.Info.Chat)
     currentTime := time.Now().In(loc)

     captionBuilder := &strings.Builder{}
     fmt.Fprintf(captionBuilder, "📋 *%s BOT MENU* (Gagal memuat gambar)\n\n", strings.ToUpper(cfg.BotName))
     fmt.Fprintf(captionBuilder, "*Waktu Server*: %s WIB\n\n", currentTime.Format("15:04:05"))
     // fmt.Fprintf(captionBuilder, "Berikut daftar fitur yang tersedia:\n") // Removed old header

     // Group commands by category (same logic as sendMenuWithImage)
     allCommands := handlers.GetAllCommands()
     commandsByCategory := make(map[string][]handlers.Command)
     for _, cmd := range allCommands {
     category := cmd.Category()
     if category == "" {
         category = "Lainnya"
     }
     commandsByCategory[category] = append(commandsByCategory[category], cmd)
     }

     categories := make([]string, 0, len(commandsByCategory))
     for category := range commandsByCategory {
     categories = append(categories, category)
     }
     // sort.Strings(categories) // Optional sort

     for _, category := range categories {
     fmt.Fprintf(captionBuilder, "\n*╭─「 %s 」*\n", strings.ToUpper(category))
     for _, cmd := range commandsByCategory[category] {
         fmt.Fprintf(captionBuilder, "*│* `%s%s` - %s\n", cfg.CommandPrefix, cmd.Name(), cmd.Description())
     }
     fmt.Fprintf(captionBuilder, "*╰──────────*\n") // Use simple line for bottom border (Fallback)
     }

     menuText := captionBuilder.String()

    // Prepare mentioned JIDs (same logic as sendMenuWithImage)
     mentionedJIDs := []string{}
     for _, owner := range cfg.Owners {
     mentionedJIDs = append(mentionedJIDs, owner)
     }

     contextInfo := &waProto.ContextInfo{
     StanzaID:      proto.String(msg.Info.ID),
     Participant:   proto.String(msg.Info.Sender.String()),
     QuotedMessage: msg.Message, // Quote pesan asli
     MentionedJID:  mentionedJIDs, // Add mentioned JIDs here
     }

     _, err := client.SendMessage(context.Background(), msg.Info.Chat, &waProto.Message{
     ExtendedTextMessage: &waProto.ExtendedTextMessage{
         Text:        proto.String(menuText),
         ContextInfo: contextInfo,
     },
     })
     if err != nil {
     // Return the error
     return fmt.Errorf("gagal kirim menu teks fallback: %w", err)
     } else {
     client.Log.Infof("Menu teks fallback terkirim ke %s!", msg.Info.Chat)
     return nil // Success
     }
}

