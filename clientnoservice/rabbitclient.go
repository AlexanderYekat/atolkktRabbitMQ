package main

import (
	fptr10 "clientrabbit/fptr"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fptr, err := fptr10.NewSafe()
	if err != nil {
		fmt.Println("Ошибка инициализации драйвера ККТ атол")
		fmt.Println(err)
		return
	}
	fmt.Println(fptr.Version())
	//jsonAnswer, err := sendComandeAndGetAnswerFromKKT(fptr, "{\"type\": \"openShift\"}")
	//failOnError(err, "Failed KKT")
	//fmt.Println(jsonAnswer)
	defer fptr.Destroy()
	//return

	//conn, err := amqp.Dial("amqp://guest:guest@80.87.106.246:5672/")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	qTask, err := ch.QueueDeclare(
		"tasks", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	qAnsw, err := ch.QueueDeclare(
		"answers", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	fmt.Println(qAnsw.Name)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		qTask.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			f, err := os.OpenFile("c:\\share\\rabbitlog.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				panic(err)
			}
			log.SetOutput(os.Stdout)
			log.Printf("Received a message: %s", d.Body)
			log.SetOutput(f)
			log.Printf("Received a message: %s", d.Body)
			f.Close()
			jsonAnswer, err := sendComandeAndGetAnswerFromKKT(fptr, string(d.Body))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = ch.PublishWithContext(ctx,
				"",         // exchange
				qAnsw.Name, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(jsonAnswer),
				})
			cancel()
			failOnError(err, "Failed to publish a message")
		}
	}()

	log.Printf(" [*] Waiting for messages v1.3(no service). To exit press CTRL+C")
	<-forever
}

func sendComandeAndGetAnswerFromKKT(fptr *fptr10.IFptr, comJson string) (string, error) {
	//return "", nil
	fptr.SetParam(fptr10.LIBFPTR_PARAM_JSON_DATA, comJson)
	fptr.ValidateJson()
	//fptr.ProcessJson()
	//result := fptr.GetParamString(fptr10.LIBFPTR_PARAM_JSON_DATA)
	result := "{\"result\": \"all ok\"}"
	return result, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
