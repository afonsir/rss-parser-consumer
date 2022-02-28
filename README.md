# Start App

```bash
RABBITMQ_URI='amqp://<USER>:<PASSWORD>@localhost:5672' \
RABBITMQ_QUEUE='store_feed_entries' \
MONGODB_URI='mongodb://<USER>:<PASSWORD>@localhost:27017/test?authSource=admin' \
MONGODB_DATABASE=demo \
go run main.go
```
