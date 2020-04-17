package main

import (
    "utils"
    "context"
)


func main() {


    c := make(chan []byte)

    kafka := utils.NewKafkaWriter([]string{"localhost:9092"},"test",1500,400,3,1,true)
    kafka.MessageHandler(c)

    server := utils.NewListenServer("udp",":12345")
    server.Run(context.Background(),c)

}
