package main

import (
	"fmt"
	"isayevapps/sunposition/engine"
	"log"

	tm "github.com/and3rson/telemux"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const token = "5573501671:AAGKD4EPV7h3qH7wYxNOWVrFU0bIcH97j6w"

var bot, _ = tgbotapi.NewBotAPI(token)

func main() {
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	mux := tm.NewMux().
		AddHandler(tm.NewConversationHandler(
			"getting_params_dialog",
			tm.NewLocalPersistence(),
			tm.StateMap{
				"": {tm.NewHandler(tm.IsCommandMessage("start"), func(u *tm.Update) {
					text := "Привет " + u.Message.From.FirstName + "!\nЯ помогу тебе найти солнце.\nДля начала я должен узнать кое-что о тебе.\nПродолжить?"
					msg := tgbotapi.NewMessage(u.Message.Chat.ID, text)
					msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Да"),
							tgbotapi.NewKeyboardButton("Нет"),
						),
					)
					bot.Send(msg)
					u.PersistenceContext.SetState("confirm_continuation")
				}),
				},
				"confirm_continuation": {
					tm.NewHandler(tm.HasText(), func(u *tm.Update) {
						if u.Message.Text == "Да" {
							msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Отправьте свою геолокацию")
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
							bot.Send(msg)
							u.PersistenceContext.SetState("send_location")
						} else {
							msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Ладно, поработаем в следующий раз")
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
							bot.Send(msg)
							u.PersistenceContext.SetState("")
						}
					}),
				},
				"send_location": {
					tm.NewHandler(tm.HasLocation(), func(u *tm.Update) {
						data := u.PersistenceContext.GetData()
						data["latitude"] = u.Message.Location.Latitude
						data["longitude"] = u.Message.Location.Longitude
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Отправьте дату в формате дд.мм.гггг"))
						u.PersistenceContext.SetData(data)
						u.PersistenceContext.SetState("send_date")
					}),
					tm.NewHandler(tm.Not(tm.IsCommandMessage("cancel")), func(u *tm.Update) {
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Извините, не узнаю локацию. Попробуйте ещё раз"))
					}),
				},
				"send_date": {
					tm.NewHandler(tm.HasRegex(`^\d{2}.\d{2}.\d{4}$`), func(u *tm.Update) {
						data := u.PersistenceContext.GetData()
						data["date"] = u.Message.Text
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Отправьте время в формате чч:мм"))
						u.PersistenceContext.SetData(data)
						u.PersistenceContext.SetState("send_time")
					}),
					tm.NewHandler(tm.Not(tm.IsCommandMessage("cancel")), func(u *tm.Update) {
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Извините, не узнаю дату. Попробуйте ещё раз"))
					}),
				},
				"send_time": {
					tm.NewHandler(tm.HasRegex(`^\d{2}:\d{2}$`), func(u *tm.Update) {
						data := u.PersistenceContext.GetData()
						data["time"] = u.Message.Text
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Отправьте часовой пояс"))
						u.PersistenceContext.SetData(data)
						u.PersistenceContext.SetState("send_gmt")
					}),
					tm.NewHandler(tm.Not(tm.IsCommandMessage("cancel")), func(u *tm.Update) {
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Извините, не узнаю время. Попробуйте ещё раз"))
					}),
				},
				"send_gmt": {
					tm.NewHandler(tm.HasRegex(`^(\+|-){0,1}\d{1,}\.{0,1}\d{0,}$`), func(u *tm.Update) {
						data := u.PersistenceContext.GetData()
						data["gmt"] = u.Message.Text
						sunPositon, err := engine.GetSunPosition(
							data["latitude"].(float64),
							data["longitude"].(float64),
							data["date"].(string),
							data["time"].(string),
							data["gmt"].(string),
						)
						if err != nil {
							bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, err.Error()+"\nЧтобы начать заново нажмите /start"))
						} else {
							result := fmt.Sprintf("Азимут: %v\nВысота: %v\n\nПараметры:\nШирота : %v\nДолгота: %v\nДата: %v\nВремя: %v\nЧасововой пояс: %v",
								sunPositon.Azimuth,
								sunPositon.Altitude,
								data["latitude"].(float64),
								data["longitude"].(float64),
								data["date"].(string),
								data["time"].(string),
								data["gmt"].(string),
							)
							bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, result))
						}
						u.PersistenceContext.ClearData()
						u.PersistenceContext.SetState("")
					}),
					tm.NewHandler(tm.Not(tm.IsCommandMessage("cancel")), func(u *tm.Update) {
						bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Извините, не узнаю часовой пояс. Попробуйте ещё раз"))
					}),
				},
			},
			[]*tm.Handler{
				tm.NewHandler(tm.IsCommandMessage("cancel"), func(u *tm.Update) {
					u.PersistenceContext.ClearData()
					u.PersistenceContext.SetState("")
					bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Отменено.\nЧтобы начать заново нажмите /start"))
				}),
			},
		))

	for update := range updates {
		mux.Dispatch(bot, update)
	}
}
