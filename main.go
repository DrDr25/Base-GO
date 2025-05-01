package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    "whatsappbot/config" // Import config package
    "whatsappbot/handlers" // Import handlers package

    // Import command packages to trigger their init() for registration
    _ "whatsappbot/commands/menu"
    _ "whatsappbot/commands/ping" // Import ping command package
)

// Simpan waktu mulai bot secara global
var startTime = time.Now()

func main() {
    // Load configuration first
     err := config.LoadConfig()
     if err != nil {
     // Use fmt.Printf for fatal startup errors before logger is ready
     fmt.Printf("FATAL: Gagal memuat konfigurasi: %v\n", err)
     return
     }

     cfg := config.GetConfig() // Get the loaded config

    // Setup logging based on config
     clientLog := waLog.Stdout("Client", cfg.LogLevel, true)
     dbLog := waLog.Stdout("Database", cfg.LogLevel, true)

     clientLog.Infof("Konfigurasi dimuat: %+v", cfg)

    // Setup database
     container, err := sqlstore.New("sqlite3", "file:whatsmeow.db?_foreign_keys=on", dbLog)
     if err != nil {
     clientLog.Errorf("Gagal buka database: %v", err)
     return
     }
     deviceStore, err := container.GetFirstDevice()
     if err != nil {
     clientLog.Errorf("Gagal ambil device: %v", err)
     return
     }

    // Buat client WhatsApp
     client := whatsmeow.NewClient(deviceStore, clientLog)

    // Tambahkan event handler
     client.AddEventHandler(func(evt interface{}) {
     RegisterEventHandlers(client, evt, cfg) // Pass config to event handlers
     })

    // Koneksi atau pairing (gunakan nomor dari config)
     ConnectOrPair(client, cfg.PairingPhone)

    // Tunggu sinyal interrupt (Ctrl+C)
     WaitForInterrupt(client)
}

// Fungsi untuk menangani event
func RegisterEventHandlers(client *whatsmeow.Client, evt interface{}, cfg config.Config) {
     switch v := evt.(type) {
     case *events.Message:
     // Log pesan masuk (Reverted to original fmt.Println version)
     handlers.PrintMessageLog(v)

     // Proses command jika pesan adalah teks dan sesuai format
     // Pastikan pesan tidak kosong dan bukan dari bot itu sendiri (opsional)
     if v.Message.GetConversation() != "" && !v.Info.IsFromMe {
         handlers.ProcessCommand(client, v, cfg, startTime)
     }
     // TODO: Handle other message types (image, video, etc.) if needed
     case *events.Connected:
     client.Log.Infof("Event: Terhubung!")
     case *events.Disconnected:
     client.Log.Infof("Event: Terputus!")
     // TODO: Handle other events as needed
     }
}

// Fungsi untuk koneksi atau pairing
func ConnectOrPair(client *whatsmeow.Client, phone string) {
     var err error
     // Jika belum ada ID, lakukan pairing
     if client.Store != nil && client.Store.ID == nil {
     client.Log.Infof("Belum ada device tersimpan, mulai proses pairing...")
     if phone == "" {
         client.Log.Errorf("Nomor telepon untuk pairing kosong di config.yaml")
         os.Exit(1) // Keluar jika nomor pairing kosong
     }
     err = client.Connect()
     if err != nil {
         client.Log.Errorf("Gagal konek untuk pairing: %v", err)
         return
     }

     code, err := client.PairPhone(phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
     if err != nil {
         client.Log.Errorf("Gagal buat kode pairing: %v", err)
         return
     }
     client.Log.Infof("Kode pairing: %s", code)
     client.Log.Infof("Silakan scan kode QR atau masukkan kode di perangkat tertaut WhatsApp Anda.")
     } else {
     client.Log.Infof("Device sudah tersimpan, mencoba konek ulang...")
     err = client.Connect()
     if err != nil {
         client.Log.Errorf("Gagal konek ulang: %v", err)
         return
     }
     }
}

// Fungsi untuk menunggu sinyal interrupt
func WaitForInterrupt(client *whatsmeow.Client) {
     c := make(chan os.Signal, 1)
     signal.Notify(c, os.Interrupt, syscall.SIGTERM)
     <-c
     client.Log.Infof("Menerima sinyal interrupt, disconnecting...")
     client.Disconnect()
}

