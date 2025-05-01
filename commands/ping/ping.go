package ping

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/host"
    "github.com/shirou/gopsutil/v3/mem"

    "go.mau.fi/whatsmeow"
    waProto "go.mau.fi/whatsmeow/binary/proto"
    "go.mau.fi/whatsmeow/types/events"
    "google.golang.org/protobuf/proto"

    "whatsappbot/config"
    "whatsappbot/handlers"
)

// Define a struct for the ping command
type pingCommand struct{}

// init registers the ping command
func init() {
    // Register an instance of pingCommand
     handlers.RegisterCommand(&pingCommand{})
}

// Name returns the command name
func (cmd *pingCommand) Name() string {
    return "ping"
}

// Description returns the command description
func (cmd *pingCommand) Description() string {
    return "Cek kecepatan respon bot dan info sistem"
}

// Aliases returns command aliases
func (cmd *pingCommand) Aliases() []string {
    return []string{"p"}
}

// Category returns the command category
func (cmd *pingCommand) Category() string {
    return "Info" // Assigning ping to the Info category
}

// Execute runs the command logic
func (cmd *pingCommand) Execute(client *whatsmeow.Client, msg *events.Message, args []string, cfg config.Config, startTime time.Time) error {
    // Calculate uptime
     uptime := time.Since(startTime)

    // Get system info
     hostInfo, errHost := host.Info()
     cpuInfo, errCPU := cpu.Info()
     memInfo, errMem := mem.VirtualMemory()

    // Prepare info strings with error handling
     osStr := "N/A"
     if errHost == nil {
     osStr = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
     }

     cpuStr := "N/A"
     if errCPU == nil && len(cpuInfo) > 0 {
     cpuStr = fmt.Sprintf("%s (%d cores)", cpuInfo[0].ModelName, cpuInfo[0].Cores)
     }

     memTotalMB := uint64(0)
     memFreeMB := uint64(0)
     if errMem == nil {
     memTotalMB = memInfo.Total / 1024 / 1024
     memFreeMB = memInfo.Available / 1024 / 1024 // Use Available for a more realistic free memory value
     }
     memStr := fmt.Sprintf("%d MB/%d MB", memFreeMB, memTotalMB)
     freeMemStr := fmt.Sprintf("%d MB", memFreeMB)

    // Format uptime
     uptimeStr := formatPingUptime(uptime)

    // Get current time and date (WIB)
     loc, _ := time.LoadLocation("Asia/Jakarta") // Load WIB timezone
     now := time.Now().In(loc)
     timeStr := now.Format("15:04:05")
     dateStr := now.Format("02/01/2006")

    // Build the response message
     responseBuilder := &strings.Builder{}
     fmt.Fprintf(responseBuilder, "*🏓 PONG!*\n\n")
     fmt.Fprintf(responseBuilder, "*System Info:*\n")
     fmt.Fprintf(responseBuilder, "• OS: %s\n", osStr)
     fmt.Fprintf(responseBuilder, "• CPU: %s\n", cpuStr)
     fmt.Fprintf(responseBuilder, "• Memory: %s\n", memStr)
     fmt.Fprintf(responseBuilder, "• Free Memory: %s\n", freeMemStr)
     fmt.Fprintf(responseBuilder, "• Uptime: %s\n", uptimeStr)
     fmt.Fprintf(responseBuilder, "• Time: %s\n", timeStr)
     fmt.Fprintf(responseBuilder, "• Date: %s\n", dateStr)

     responseText := responseBuilder.String()

    // Send the response, quoting the original message
     _, err := client.SendMessage(context.Background(), msg.Info.Chat, &waProto.Message{
     ExtendedTextMessage: &waProto.ExtendedTextMessage{
         Text: proto.String(responseText),
         ContextInfo: &waProto.ContextInfo{
	 StanzaID:      proto.String(msg.Info.ID),
	 Participant:   proto.String(msg.Info.Sender.String()),
	 QuotedMessage: msg.Message,
         },
     },
     })

     if err != nil {
     // Return the error to be logged by the caller (ProcessCommand)
     return fmt.Errorf("gagal mengirim balasan ping: %w", err)
     }
     return nil // Success
}

// formatPingUptime formats duration for the ping command uptime
func formatPingUptime(d time.Duration) string {
     d = d.Round(time.Second)
     h := d / time.Hour
     d -= h * time.Hour
     m := d / time.Minute
     d -= m * time.Minute
     s := d / time.Second
     return fmt.Sprintf("%d hours, %d minutes, %d seconds", h, m, s)
}

