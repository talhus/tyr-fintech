package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/iamtbay/tyr-fintech/internal/dto"
)

var WebHookQueue = make(chan *dto.TransactionWebhookEvent, 100)

func StartWebhookWorker() {
	log.Println("Webhook worker has started, tasks are awaiting")
	for tx := range WebHookQueue {
		sendWebhook(tx)
	}
}

func sendWebhook(tx *dto.TransactionWebhookEvent) {
	//could be dynamic from db for user.
	webhookUrl := "https://webhook.site/410c1355-7fd9-4be1-b2a6-2f2336fbff21"
	//convert to json
	jsonData, _ := json.Marshal(tx)

	//http req
	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Webhook error] Process ID %s couldn't sent %v", tx.TransactionID, err)
		return
	}

	defer resp.Body.Close()
	fmt.Printf("[Webhook success] Process ID %s sent!", tx.TransactionID)
}
