package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/it1804/kafka-bridge/common/config"
	"github.com/it1804/kafka-bridge/services"
	"github.com/it1804/kafka-bridge/services/http"
	"github.com/it1804/kafka-bridge/services/kafka"
	"github.com/it1804/kafka-bridge/services/stat"
	"github.com/it1804/kafka-bridge/services/udp"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
)

func main() {

	c := flag.String("c", "config.json", "Specify the configuration file.")
	flag.Parse()
	jsonf, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Fatalf("Error config file: %v\n", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(jsonf, &data); err != nil {
		log.Fatalf("Error config file: %v\n", err)
	}
	conf := &config.ServiceList{}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &conf,
	})

	decoder.Decode(data)

	ctxWithCancel, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	wg := &sync.WaitGroup{}

	var statService services.StatService

	if len(conf.Stat.Listen) > 0 {
		statService = stat.NewStatService(ctxWithCancel, wg, &conf.Stat)
	}

	for i := range conf.Services {
		if conf.Services[i].Type == "http" {
			statService.Watch(http.NewHttpService(ctxWithCancel, wg, &conf.Services[i]))
		} else if conf.Services[i].Type == "udp" {
			statService.Watch(udp.NewUdpService(ctxWithCancel, wg, &conf.Services[i]))
		} else if conf.Services[i].Type == "kafka_consumer" {
			statService.Watch(kafka.NewKafkaService(ctxWithCancel, wg, &conf.Services[i]))
		} else {
			log.Fatalf("Unknown service name: %v\n", conf.Services[i].Type)
		}
	}

	select {
	case <-sig:
		cancel()
		wg.Wait()
	}
	log.Printf("kafka-bridge exited\n")

}
