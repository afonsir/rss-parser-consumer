package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`

	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`

	Title string `xml:"title"`
}

type Request struct {
	URL string `json:"url"`
}

var client *mongo.Client
var ctx context.Context

func init() {
	ctx = context.Background()

	client, _ = mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGODB_URI")))

	log.Println("Connected to MongoDB")
}

func main() {
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}

	defer amqpConnection.Close()
	channelAmqp, _ := amqpConnection.Channel()
	defer channelAmqp.Close()

	forever := make(chan bool)

	msgs, err := channelAmqp.Consume(
		os.Getenv("RABBITMQ_QUEUE"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var request Request
			json.Unmarshal(d.Body, &request)
			log.Println("RSS URL:", request.URL)

			entries, _ := GetFeedEntries(request.URL)
			collection := client.Database(os.Getenv("MONGODB_DATABASE")).Collection("recipes")

			for _, entry := range entries[2:] {
				collection.InsertOne(ctx, bson.M{
					"title":     entry.Title,
					"thumbnail": entry.Thumbnail.URL,
					"url":       entry.Link.Href,
				})
			}
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")

	<-forever
}

func GetFeedEntries(url string) ([]Entry, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(
		"User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	byteValue, _ := ioutil.ReadAll(resp.Body)
	var feed Feed
	xml.Unmarshal(byteValue, &feed)
	return feed.Entries, nil
}
