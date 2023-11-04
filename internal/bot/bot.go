package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/AndreyAD1/helsinki-guide/internal/bot/configuration"
	"github.com/AndreyAD1/helsinki-guide/internal/bot/handlers"
	"github.com/AndreyAD1/helsinki-guide/internal/bot/services"
	"github.com/AndreyAD1/helsinki-guide/internal/infrastructure/repositories"
	"github.com/AndreyAD1/helsinki-guide/internal/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	bot             *tgbotapi.BotAPI
	handlers        handlers.HandlerContainer
	shutdownFuncs   []func()
	tgUpdateTimeout int
}

func NewServer(ctx context.Context, config configuration.StartupConfig) (*Server, error) {
	bot, err := tgbotapi.NewBotAPI(config.BotAPIToken)
	if err != nil {
		return nil, fmt.Errorf("can not connect to the Telegram API: %w", err)
	}
	bot.Debug = false
	dbpool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to create a connection pool: DB URL '%s': %w",
			config.DatabaseURL,
			err,
		)
	}
	if err := dbpool.Ping(ctx); err != nil {
		logMsg := fmt.Sprintf(
			"unable to connect to the DB '%v'", 
			config.DatabaseURL,
		)
		slog.ErrorContext(ctx, logMsg, slog.Any(logger.ErrorKey, err))
		return nil, fmt.Errorf("%v: %w", logMsg, err)
	}
	buildingRepo := repositories.NewBuildingRepo(dbpool)
	buildingService := services.NewBuildingService(buildingRepo)
	handlerContainer := handlers.NewCommandContainer(bot, buildingService)
	server := Server{
		bot,
		handlerContainer,
		[]func(){dbpool.Close},
		config.TGUpdateTimeout,
	}
	return &server, nil
}

func (s *Server) Shutdown() {
	for _, f := range s.shutdownFuncs {
		f()
	}
}

func (s *Server) RunBot(ctx context.Context) error {
	if err := s.setBotCommands(ctx); err != nil {
		return fmt.Errorf("can not set bot commands: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = s.tgUpdateTimeout

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	updates := s.bot.GetUpdatesChan(u)

	go s.receiveUpdates(ctx, updates)

	fmt.Println("Start listening for updates")
	<-ctx.Done()
	return nil
}

func (s *Server) setBotCommands(ctx context.Context) error {
	commands := []tgbotapi.BotCommand{}
	for commandName, handler := range s.handlers.HandlersPerCommand {
		command := tgbotapi.BotCommand{
			Command:     commandName,
			Description: handler.Description,
		}
		commands = append(commands, command)
	}
	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
	result, err := s.bot.Request(setCommandsConfig)
	if err != nil {
		slog.ErrorContext(
			ctx,
			fmt.Sprintf(
				"can not set commands: command config: %v",
				setCommandsConfig,
			),
			slog.Any(logger.ErrorKey, err),
		)
		return fmt.Errorf("can not make a request to set commands: %w", err)
	}
	if !result.Ok {
		err = fmt.Errorf(result.Description)
		slog.ErrorContext(
			ctx,
			fmt.Sprintf(
				"can not set commands: an error code '%v': a response body '%s'",
				result.ErrorCode,
				result.Result,
			),
			slog.Any(logger.ErrorKey, err),
		)
		return err
	}

	return nil
}

func (s *Server) receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			s.handleUpdate(ctx, update)
		}
	}
}

func (s *Server) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		s.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		s.handleButton(ctx, update.CallbackQuery)
	}
}

func (s *Server) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	user := message.From

	if user == nil {
		return
	}
	handler, ok := s.handlers.GetHandler(message.Command())
	if !ok {
		answer := fmt.Sprintf(
			"Unfortunately, I don't understand this message: %s",
			message.Text,
		)
		responseMsg := tgbotapi.NewMessage(message.Chat.ID, answer)
		if _, err := s.bot.Send(responseMsg); err != nil {
			logMsg := fmt.Sprintf(
				"can not send a message to %v: %v",
				message.Chat.ID, 
				answer,
			)
			slog.WarnContext(ctx, logMsg, slog.Any(logger.ErrorKey, err))
		}
		return
	}
	handler.Function(s.handlers, ctx, message)
}

func (s *Server) handleButton(ctx context.Context, query *tgbotapi.CallbackQuery) {
	var queryData handlers.Button
	if err := json.Unmarshal([]byte(query.Data), &queryData); err != nil {
		slog.WarnContext(
			ctx, 
			fmt.Sprintf("unexpected callback data %v", query), 
			slog.Any(logger.ErrorKey, err),
		)
		return
	}
	handler, ok := s.handlers.GetButtonHandler(queryData.Name)
	if !ok {
		logMsg := fmt.Sprintf(
			"the unexpected button name %v from the chat %v: initial message %v",
			queryData,
			query.Message.Chat.ID,
			query.Message.MessageID,
		)
		slog.WarnContext(ctx, logMsg)
		return
	}
	handler(s.handlers, ctx, query)
}
