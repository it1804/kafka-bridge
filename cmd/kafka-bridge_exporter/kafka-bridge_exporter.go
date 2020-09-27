package main

import (
	"encoding/json"
	"flag"
	"github.com/it1804/kafka-bridge/config"
	"github.com/it1804/kafka-bridge/exporter"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	c := flag.String("c", "metrics.json", "Specify the configuration file.")
	flag.Parse()
	jsonf, err := ioutil.ReadFile(*c)
	if err != nil {
		log.Fatalf("Error config file: %v\n", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(jsonf, &data); err != nil {
		log.Fatalf("Error config file: %v\n", err)
	}
	conf := &config.ExporterServiceList{}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &conf,
	})

	decoder.Decode(data)

	if len(conf.Metrics.Listen) == 0 {
		conf.Metrics.Listen = ":2112"
	}

	if len(conf.Metrics.Path) == 0 {
		conf.Metrics.Path = "/metrics"
	}

	//    b, _ := json.MarshalIndent(conf, "", "  ")
	//    log.Printf("Running config: %s\n", string(b))

	kb := exporter.NewCollector(conf)
	prometheus.MustRegister(kb)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>Kafka-bridge exporter</title></head>
<body>
<h1>Kafka-bridge exporter</h1>
<p><a href='` + conf.Metrics.Path + `'>Metrics</a></p>
</body>
</html>
`))
	})

	http.Handle(conf.Metrics.Path, promhttp.Handler())
	log.Fatal(http.ListenAndServe(conf.Metrics.Listen, nil))
}
