package main

import (
	"context"
	"log"
	"os"
	"time"
	fptr10 "clientrabbit/fptr"

	amqp "github.com/rabbitmq/amqp091-go"
	//svc "golang.org/x/sys/windows/svc"
	//eventlog "golang.org/x/sys/windows/svc/eventlog"
	service "github.com/kardianos/service"
)

type myProgram struct{}

func (p *myProgram) Start(s service.Service) error {
	// Operation executed when starting the service
	log.Println("Service start...")
	go p.run()
	return nil
}

func (p *myProgram) run() {
	// Write your service logic code here
	log.Println("Service running...")

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
	failOnError(err, "Failed to declare a queue tasks")
	qAnsw, err := ch.QueueDeclare(
		"answers", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue answers")

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
			jsonAnswer, err := SendComandeAndGetAnswerFromKKT(d.Body)
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

			log.SetOutput(os.Stdout)
			log.Printf("Service Received a message: %s", d.Body)
			log.SetOutput(f)
			log.Printf("Service Received a message: %s", d.Body)
			f.Close()
		}
	}()

	log.Printf(" [*] Waiting for messages v1.3. To exit press CTRL+C")
	<-forever
}

func SendComandeAndGetAnswerFromKKT(fptr *fptr10.IFptr, comJson string) (string, error) {
	fptr.setParam(fptr.LIBFPTR_PARAM_JSON_DATA, comJson)
	fptr.processJson()
	result := fptr.GetParamString(fptr10.LIBFPTR_PARAM_JSON_DATA)
	return result, nil
}

func (p *myProgram) Stop(s service.Service) error {
	// Operation executed when stopping the service
	log.Println("Service stopped.")
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "MyService",
		DisplayName: "My Go Service",
		Description: "This is a Windows service written in Go.",
	}
	prg := &myProgram{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	// Run in a service manner
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
