package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	config := GetConfig()

	if config.Environment == "production" {
		// Need to spin up a web server for Google Cloud Run to use for health checks
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		go func() {
			if err := http.ListenAndServe(":8080", nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	session, err := discordgo.New("Bot " + config.AuthToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Bot session created")

	err = session.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Listen for messages in channel
	session.AddHandler(messageCreate)

	log.Printf("Bot is running as %s\n.", session.State.User.Username)
	<-make(chan struct{})
	session.Close()
	return
}

// Runs on every message created in the discord server
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, config *Config) {
	log.Printf("Message received")

	// Ignore all messages created by the bot itself and messages not in MOD_CHANNEL
	if (m.Author == s.State.User) || (m.ChannelID != config.ModChannelId) {
		return
	}

	// Check for the last message in the channel from this user
	lastMsg := getPreviousUserMessage(s, m.ChannelID, m.Author.ID)
	if lastMsg == nil {
		log.Printf("No previous message found for %s\n", m.Author.Username)
		return
	}
	log.Printf("Last message time: %s\n", lastMsg.Timestamp)

	// Check if the last message was within the cooldown period
	timeDiff := m.Timestamp.Sub(lastMsg.Timestamp)
	timeLeft := getTimeLeft(timeDiff, config.CoolDownHours)
	log.Printf("Message from %s is %f hours old\n", m.Author.Username, timeDiff.Hours())
	if timeLeft == 0 {
		log.Printf("No cooldown for %s\n", m.Author.Username)
		return
	}

	// Delete the message
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Printf("Error deleting message: %s\n", err)
		return
	}
	log.Printf("Deleted message from %s\n", m.Author.Username)

	// Ping the user in logs channel
	_, err = s.ChannelMessageSend(config.LogChannelId, m.Author.Mention()+" The Keeper rejects your proverb. You must wait "+formatDuration(timeLeft)+" before posting again.")
	if err != nil {
		log.Printf("Error sending message: %s\n", err)
		return
	}
	log.Printf("Sent message to %s\n", config.LogChannelId)
}

func getPreviousUserMessage(s *discordgo.Session, channelID string, userID string) *discordgo.Message {
	// Get the last 100 messages from the channel that are from the user
	msgs, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Printf("Error getting messages: %s\n", err)
		return nil
	}

	// Loop through the messages and find the first one that is from the user
	for _, msg := range msgs[1:] {
		if msg.Author.ID == userID {
			return msg
		}
	}
	return nil
}

// Returns time left in seconds
func getTimeLeft(timeDiff time.Duration, cooldown int) float64 {
	cd := float64(cooldown) * 60 * 60 // cooldown time in seconds
	td := (timeDiff).Seconds()

	if td < cd {
		return cd - td
	}
	return 0
}

// given an duration d in seconds, returns a string formatted as "x hours", "x minutes", or "x seconds"
func formatDuration(d float64) string {
	hours := d / 60 / 60
	minutes := d / 60
	seconds := d

	// round to 1 decimal place if hours is a float
	if hours >= 1 {
		if hours*10.0 == float64(int(hours*10.0)) {
			return fmt.Sprintf("%.0f hours", hours)
		}
		return fmt.Sprintf("%.1f hours", hours)
	}
	if minutes >= 1 {
		return fmt.Sprintf("%.0f minutes", minutes)
	}
	return fmt.Sprintf("%.0f seconds", seconds)
}
