package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var discordWebhookURL string

// Replace with your Discord webhook URL

// Define structs for Bitbucket webhook payload
type WebhookPayload struct {
	Push struct {
		Changes []struct {
			New struct {
				Target struct {
					Hash    string `json:"hash"`
					Message string `json:"message"`
					Links   struct {
						HTML struct {
							Href string `json:"href"`
						} `json:"html"`
					} `json:"links"`
					Author struct {
						User struct {
							DisplayName string `json:"display_name"`
						} `json:"user"`
					} `json:"author"`
				} `json:"target"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleWebhook processes incoming Bitbucket webhook requests and sends information to Discord
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload

	// Read and decode JSON payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Extract the required information
	if len(payload.Push.Changes) > 0 {
		change := payload.Push.Changes[0] // Only the first change

		committer := change.New.Target.Author.User.DisplayName
		projectName := payload.Repository.Name
		commitHash := change.New.Target.Hash
		commitURL := change.New.Target.Links.HTML.Href
		commitMessage := change.New.Target.Message

		// Format the message for Discord
		discordMessage := fmt.Sprintf(
			"🚀 **New Commit:** `%s`\n"+
				"👤 **Committer:** %s\n"+
				"🔑 **Hash:** `%s`\n"+
				"📝 **Message:** %s\n"+
				"🔗 **URL:** [View Commit](%s)",
			projectName, committer, commitHash, commitMessage, commitURL,
		)

		// Send message to Discord
		err = sendDiscordMessage(discordMessage)
		if err != nil {
			http.Error(w, "Failed to send message to Discord", http.StatusInternalServerError)
			return
		}

		// Respond to the webhook sender
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Webhook processed and sent to Discord"))
	} else {
		http.Error(w, "No changes in push event", http.StatusBadRequest)
	}
}

// sendDiscordMessage sends a message to a Discord channel using a webhook URL
func sendDiscordMessage(message string) error {
	_ = godotenv.Load()

	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookURL == "" {
		log.Fatal("DISCORD WEBHOOK URL is required")
	}

	// Prepare the payload for Discord
	payload := map[string]string{"content": message}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the HTTP request to Discord webhook URL
	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send message, status: %s", resp.Status)
	}

	return nil
}
