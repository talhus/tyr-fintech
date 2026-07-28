package notifications

import (
	"encoding/json"
	"log"
)

type NotificationPayload struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type NotificationEvent struct {
	UserID      string
	TargetEmail string
	Title       string
	Message     string
	Type        string
}

type NotificationService struct {
	hub        *Hub
	emailQueue chan *NotificationEvent
}

func NewNotificationService(hub *Hub) *NotificationService {
	ns := &NotificationService{
		hub:        hub,
		emailQueue: make(chan *NotificationEvent, 100),
	}
	go ns.startEmailWorker() //start worker

	return ns
}

func (ns *NotificationService) NotifyUser(event *NotificationEvent) {
	payload, _ := json.Marshal(NotificationPayload{
		Title:   event.Title,
		Message: event.Message,
		Type:    event.Type,
	})

	isOnline := ns.hub.SendToUser(event.UserID, string(payload))
	if isOnline {
		log.Printf("Notification sent to user %s", event.UserID)
	} else {
		log.Printf("User %s offline.", event.UserID)
	}

	//ASYNC EMAIL
	select {
	case ns.emailQueue <- event:
	default:
		log.Printf("Email queue is full. Skipped for user %s", event.UserID)
	}
}

func (ns *NotificationService) startEmailWorker() {
	log.Println("Email worker started to work on background")

	for event := range ns.emailQueue {
		//mail integration
		//mocking mail
		log.Printf("\n===Email sent to \n %s \n Topic: %s \n Message:%s \n ===\n", event.TargetEmail, event.Title, event.Message)
	}

}
