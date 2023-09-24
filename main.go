package main

import (
	"log"

	"os"

	"github.com/joho/godotenv"

	"github.com/bwmarrin/discordgo"
)

const MOD_CHANNEL_ID = "1155592884138025101"
const LOG_CHANNEL_ID = "1155596136602677300"

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("AUTH_TOKEN")

	session, err := discordgo.New("Bot " + token)
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

// on messages created in MOD_CHANNEL, call this
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself and messages not in MOD_CHANNEL
	if (m.Author == s.State.User) || (m.ChannelID != MOD_CHANNEL_ID) {
		return
	}

	log.Printf("Message received")

	// Check for the last message in the channel from this user
	lastMsg := getPreviousUserMessage(s, m.ChannelID, m.Author.ID)
	if lastMsg == nil {
		log.Printf("No previous message found for %s\n", m.Author.Username)
		return
	}

	// If it's less than 6 hours old, delete the new message and send a warning in LOG_CHANNEL
	log.Printf("Last message time: %s\n", lastMsg.Timestamp)

	timeDiff := m.Timestamp.Sub(lastMsg.Timestamp)

	if (timeDiff).Hours() < 6 {
		log.Printf("Message from %s is %f hours old\n", m.Author.Username, timeDiff.Hours())

		// Delete the message
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			log.Printf("Error deleting message: %s\n", err)
			return
		}

		// Ping user and warn them for posting too soon
		_, err = s.ChannelMessageSend(LOG_CHANNEL_ID, m.Author.Mention()+" posted too soon after their last message")
		if err != nil {
			log.Printf("Error sending message: %s\n", err)
			return
		}

		log.Printf("Deleted message from %s\n", m.Author.Username)
	}

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

func deleteMessage(s *discordgo.Session, channelID string, messageID string) {
	err := s.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		log.Printf("Error deleting message: %s\n", err)
		return
	}
}
