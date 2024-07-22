package main

/*
#cgo pkg-config: lept tesseract
#cgo LDFLAGS: -L/opt/homebrew/lib -lleptonica -ltesseract
#cgo CFLAGS: -I/opt/homebrew/include
*/
import "C"

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/otiai10/gosseract/v2"
)

func main() {
    botToken := os.Getenv("BOT_TOKEN")
    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true
    log.Printf("Authorized on account %s", bot.Self.UserName)

 
    _, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
    if err != nil {
        log.Fatalf("Failed to delete webhook: %v", err)
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message != nil { 
            if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
                fileID := update.Message.Photo[len(update.Message.Photo)-1].FileID
                file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
                if err != nil {
                    log.Println("Failed to get file:", err)
                    continue
                }

                fileURL := file.Link(bot.Token)
                log.Println("File URL:", fileURL)

              
                resp, err := http.Get(fileURL)
                if err != nil {
                    log.Println("Failed to download file:", err)
                    continue
                }
                defer resp.Body.Close()

             
                tmpfile, err := ioutil.TempFile("", "image-*.png")
                if err != nil {
                    log.Println("Failed to create temp file:", err)
                    continue
                }
                defer os.Remove(tmpfile.Name()) 

               
                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                    log.Println("Failed to read response body:", err)
                    continue
                }

                _, err = tmpfile.Write(body)
                if err != nil {
                    log.Println("Failed to write to temp file:", err)
                    continue
                }
                tmpfile.Close()

             
                client := gosseract.NewClient()
                defer client.Close()
                
              
                client.SetLanguage("eng+rus+eng+") 

                err = client.SetImage(tmpfile.Name())
                if err != nil {
                    log.Println("Failed to set image:", err)
                    continue
                }

                text, err := client.Text()
                if err != nil {
                    log.Println("Failed to extract text:", err)
                    continue
                }

                msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
                bot.Send(msg)
            }
        }
    }
}
