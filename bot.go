package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	telegramToken := ""//Here you should put your telegram bot token
	discordToken := "Bot " + ""//And in the 2nd string your discord bot API key
	var chatGroup int64 = -1234//Telegram chat group ID
	discordChannel := ""//Discord Channel ID
	debug := true//Debug phase

	//////
	//	TELEGRAM
	//////

	tgbot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		//If the telegram bot returned some king of error, abort
		fmt.Println("error creating Telegram session,", err)
		log.Panic(err)
		return
	}
	//Set dbg to the debug value
	tgbot.Debug = debug

	////
	//	Long polling settings
	////
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tgbot.GetUpdatesChan(u)

	log.Printf("Authorized telegram bot on account %s", tgbot.Self.UserName)//ACK on telegram

	//////
	//	DISCORD
	//////

	dg, err := discordgo.New(discordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)//If the discord bot returned some king of error, abort
		log.Panic(err)
		return
	}
	//Set dbg to the debug value
	dg.Debug = debug

	log.Printf("Authorized Discord bot on account")//ACK on discord

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice.
		if m.Author.ID == s.State.User.ID {
			return
		}
		// If the message is "ping" reply with "Pong!"
		if m.Content == "/ping" {
			s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
		} else {
			//Discord -> Telegram
			//Forward all the messages from discord to a chat group
			if m.ChannelID == discordChannel {
				msg := tgbotapi.NewMessage(chatGroup, m.Author.Username+": "+m.Content)
				fmt.Println(tgbot)
				tgbot.Send(msg)
			}
		}

	})

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	//////
	//	TELEGRAM -> DISCORD
	//////

	for update := range updates {
		if update.Message != nil {
			if update.Message.Text == "/ping" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🏓 Pong!")
				msg.ReplyToMessageID = update.Message.MessageID
				tgbot.Send(msg)
			} else
			//Telegram -> Discord
			if update.Message.Chat.ID == chatGroup {
				//if the message comes from the group i wanted, forward it
				dg.ChannelMessageSend(discordChannel, (update.Message.From.FirstName + " [@" + update.Message.From.UserName + "]: " + update.Message.Text))
			}
		}
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
