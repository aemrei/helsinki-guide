package handlers

import (
	c "context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/AndreyAD1/helsinki-guide/internal/bot/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewHandler(bot *tgbotapi.BotAPI, service services.BuildingService) CommandHandlerContainer {
	handlersPerCommand := map[string]Handler{
		"start":     {CommandHandlerContainer.start, "Start the bot"},
		"help":      {CommandHandlerContainer.help, "Get help"},
		"settings":  {CommandHandlerContainer.settings, "Configure settings"},
		"addresses": {CommandHandlerContainer.getAllAdresses, "Get all available addresses"},
		"building":  {CommandHandlerContainer.getBuilding, "Get building by address"},
	}
	availableCommands := []string{}
	for command := range handlersPerCommand {
		availableCommands = append(availableCommands, "/"+command)
	}
	slices.Sort(availableCommands)
	commandsForHelp := strings.Join(availableCommands, ", ")
	return CommandHandlerContainer{service, bot, handlersPerCommand, commandsForHelp}
}

func (h CommandHandlerContainer) GetHandler(command string) (Handler, bool) {
	handler, ok := h.HandlersPerCommand[command]
	return handler, ok
}

func (h CommandHandlerContainer) SendMessage(chatId int64, msgText string) {
	msg := tgbotapi.NewMessage(chatId, msgText)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func (h CommandHandlerContainer) start(ctx c.Context, message *tgbotapi.Message) {
	startMsg := "Hello! I'm a bot that helps you to understand Helsinki better."
	h.SendMessage(message.Chat.ID, startMsg)
}

func (h CommandHandlerContainer) help(ctx c.Context, message *tgbotapi.Message) {
	helpMsg := fmt.Sprintf("Available commands: %s", h.commandsForHelp)
	h.SendMessage(message.Chat.ID, helpMsg)
}

func (h CommandHandlerContainer) settings(ctx c.Context, message *tgbotapi.Message) {
	settingsMsg := "No settings yet."
	h.SendMessage(message.Chat.ID, settingsMsg)
}

func (h CommandHandlerContainer) getAllAdresses(ctx c.Context, message *tgbotapi.Message) {
	address := message.CommandArguments()
	limit := 10
	buildings, err := h.buildingService.GetBuildingPreviews(ctx, address, limit)
	if err != nil {
		log.Printf("can not get addresses: %s", err.Error())
		h.SendMessage(message.Chat.ID, "Internal error")
		return
	}
	response := "Available building addresses and names:\n"
	items := make([]string, len(buildings))
	itemTemplate := "%v. %s - %s"
	for i, building := range buildings {
		items[i] = fmt.Sprintf(itemTemplate, i+1, building.Address, building.Name)
	}
	response = response + strings.Join(items, "\n") + "\nEnd"
	msg := tgbotapi.NewMessage(message.Chat.ID, response)
	msg.ParseMode = tgbotapi.ModeHTML

	buttonLabel := fmt.Sprintf("Next %v buildings", limit)
	button := Button{buttonLabel, "next", limit, limit}
	buttonCallbackData, err := json.Marshal(button)
	if err != nil {
		log.Printf("can not create a button %v: %v", button, err)
		return
	}
	buttonData := tgbotapi.NewInlineKeyboardButtonData(
		button.label,
		string(buttonCallbackData),
	)
	moreAddressesMenuMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonData),
	)
	msg.ReplyMarkup = moreAddressesMenuMarkup
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func (h CommandHandlerContainer) getBuilding(ctx c.Context, message *tgbotapi.Message) {
	address := message.CommandArguments()
	if address == "" {
		h.SendMessage(message.Chat.ID, "Please add an address to this command.")
		return
	}
	buildings, err := h.buildingService.GetBuildingsByAddress(ctx, address)
	if err != nil {
		log.Printf("can not get building by address '%s': %s", address, err.Error())
		h.SendMessage(message.Chat.ID, "Internal error.")
		return
	}
	userLanguage := "en"
	if user := message.From; user != nil {
		userLanguage = user.LanguageCode
	}
	items := make([]string, len(buildings))
	for i, building := range buildings {
		serializedItem, err := SerializeIntoMessage(building, userLanguage)
		if err != nil {
			log.Printf("can not serialize a building '%s': %s", address, err.Error())
			items[i] = "A building error."
			continue
		}
		items[i] = serializedItem
	}
	response := strings.Join(items, "\n\n")
	h.SendMessage(message.Chat.ID, response)
}
