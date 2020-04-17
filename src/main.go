package main

import (
    "utils"
    "context"
)


func main() {


    c := make(chan []byte)

    kafka := utils.NewKafkaWriter("localhost:9092","test")
    kafka.MessageHandler(c)

    server := utils.NewListenServer("udp",":12345")
    server.Run(context.Background(),c)

}
