package main

import (
    "utils"
    "flag"
    "os"
    "log"
    "encoding/json"
)


func main() {

    c := flag.String("c", "congig.json", "Specify the configuration file.")
    flag.Parse()
    file, err := os.Open(*c)
    if err != nil {
	log.Fatal("can't open config file: ", err)
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    Config := utils.Configuration{}
    err = decoder.Decode(&Config)
    if err != nil {
	log.Fatal("can't decode config JSON: ", err)
    }

    channel := make(chan []byte,100000)
    kafka := utils.NewKafkaWriter(Config.Kafka.Brokers,Config.Kafka.Topic,Config.Kafka.Workers)
    kafka.MessageHandler(channel)
    server := utils.NewListenServer("udp",Config.Listen.Address,Config.Listen.MaxPacketSize,Config.Listen.Workers)
    server.Listen(channel)
    server.PrintStat()
}
