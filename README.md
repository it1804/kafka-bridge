UDP, HTTP, PCAP and Kafka bridge to Kafka

go get -u github.com/it1804/kafka-bridge

# kafka-bridge

## Описание
Служит для пересылки в выделенный сервер (кластер) Apache Kafka сообщений, принимаемых по протоколам http, udp, pcap и kafka.
* Сервис http:
Принимает HTTP POST и HTTP PUT сообщения, извлекает из них тело сообщения и заголовки HTTP
* Сервис udp:
Принимает сообщения по протоколу UDP, извлекается тело сообщения и ip-адрес отправителя UDP пакета (source ip)
* Сервис pcap:
Захватывает трафик c сетевого интерфейса через библиотеку libpcap, извлекается тело сообщения и ip-адрес отправителя полученного пакета (source ip)
* Сервис kafka_consumer:
Читает сообщения из стороннего сервера kafkа, извлекается тело и заголовки сообщения

Каждый принимаемый поток описывается в настройках как отдельный сервис. Сервисов в конфигурационном файле можно указать любое количество, вне зависимости от протокола, в любой комбинации.
Для получения статистики по доставленным в кафку сообщениям можно настроить отдельный http сервис статистики, который может как отдавать статистику в формате json, так и экспортировать метрики prometheus.
Информация по всем активным сервисам обновляется один раз в секунду.

##### Примечание:
Также есть отдельная утилита [it1804/kafka-bridge_exporter](https://github.com/it1804/kafka-bridge_exporter) для отправки статистики на сервер prometheus (kafka-bridge_exporter). Эта утилита может централизованно забирать и отправлять статистику сразу с нескольких экземпляров kafka-bridge через встроенный в kafka-bridge сервер статистики.

Все примеры лежат в [samples](https://github.com/it1804/kafka-bridge/tree/master/samples)

## Конфигурация
1. Тип принимающего сервиса может быть одним из ключевых слов:
* "type":"http"
* "type":"udp"
* "type":"pcap"
* "type":"kafka_consumer"
2. Имя сервиса
"name": "service-description"
Любая строка



### Сервис http
```
"http_options":{
    "listen":"127.0.0.1:8081",
    "allowed_headers":["x-auth-token","x-real-ip", "x-event-time"],
    "mode":"json",
    "path":"/stream"
} 
```
1. listen: задаёт пару ip:port для прослушивания трафика. Опция обязательна
2. allowed_headers: Разрешённые к пересылке HTTP заголовки из исходного сообщения, принятого по протоколу HTTP POST или HTTP PUT, если они были действительно получены. По умолчанию []
3. path: Путь, по которому будeт приниматься сообщения методом POST или PUT. По умолчанию "/"
4. mode: Формат записи полученного сообщения в кафку:
* "body" - В сообщение будет записано только тело HTTP сообщения. Разрешённые HTTP заголовки будут помещены в служебные заголовки сообщения кафки.
* "base64_body" - В сообщение будет записано только тело HTTP сообщения, закодированное методом base64. Разрешённые HTTP заголовки будут помещены в служебные заголовки сообщения кафки.
* "json" - Тело HTTP сообщения будет упаковано в json строку в секцию "body", а HTTP заголовки в секцию "headers". Также разрешённые HTTP заголовки будут помещены в служебные заголовки сообщения кафки.
* "raw" - Сообщение HTTP будет записано в "сыром" виде, в том числе и все принятые HTTP заголовки. Смотрите [https://pkg.go.dev/net/http#Request.WriteProxy](https://pkg.go.dev/net/http#Request.WriteProxy). Также разрешённые HTTP заголовки будут помещены в служебные заголовки сообщения кафки.

##### Примечание:
Тестовое сообщение по HTTP POST можно, например, отправить вот так:

```curl -d 'Test message' -H "Content-Type: text/plain" https://localhost:8080/stream```


### Сервис udp
```
"udp_options":{
            "listen":"0.0.0.0:3000",
            "max_packet_size":65535,
            "base64_encode_body":false,
            "workers":2,
            "signature_bytes":{ "0":"0x7b", "1":"0x22" },
            "source_ip_header":"src-ip"
         }, 
```
1. listen: задаёт пару ip:port для прослушивания трафика. Опция обязательна
2. max_packet_size: максимальный размер UDP сообщения, который может быть принят. Пакеты, длина которых больше, будут обрезаться до указанного параметра. По умолчанию 65535
3. base64_encode_body: тело сообщения будет кодировано перед отправкой методом base64. По умолчанию false
4. workers: Количество одновременных воркеров(тредов), которые будут параллельно принимать UDP сообщения. Имеет смысл делать больше одного-двух на многопроцессорных машинах при очень больших pps (пакетов в секунду) и при потерях пакетов внутри этой машины. В большинстве случаев такие проблемы решаются тюнингом sysctl. По умолчанию 1
5. source_ip_header: В заголовки сообщения кафки будет добавлен заголовок с указанным именем, в котором будет передаваться ip-адрес источника udp дейтаграммы в виде строки.. По умолчанию заголовок не передаётся
6. signature_bytes: байтовый фильтр, который позволяет по содержимому определенных байт сообщения ограничить приём UDP сообщений, если заранее известен протокол, но неизвестны адреса отправителей
```"signature_bytes":{ "0":"0x7b", "1":"0x22" }``` например, ограничивает приём сообщений, которая начинается с ```{"```, т.е. начало json строки. Все остальные сообщения будут заблокированы и залогированы. По умолчанию {}

### Сервис pcap
```
"pcap_options": {
            "device": "eth1",
            "filter": "udp src port 1111",
            "snap_len": 65535,
            "base64_encode_body":false,
            "source_ip_header_name":"src-ip",
            "source_ip_as_partition_key":true
          },

```
1. device: задаёт устройство прослушивания трафика. Опция обязательна
2. filter: Выражения для BPF фильтра [https://www.tcpdump.org/manpages/pcap-filter.7.html](https://www.tcpdump.org/manpages/pcap-filter.7.html). Опция обязательна
3. snap_len: максимальный размер сообщения, который может быть принят через BPF фильтр. Пакеты, длина которых больше, будут обрезаться до указанного параметра. По умолчанию 65535
4. base64_encode_body: тело сообщения будет кодировано перед отправкой методом base64. По умолчанию false
5. source_ip_header: В заголовки сообщения кафки будет добавлен заголовок с указанным именем, в котором будет передаваться ip-адрес источника, извлечённый из захваченного пакета в виде строки. По умолчанию заголовок не передаётся
6. source_ip_as_partition_key: ip-адрес источника будет использован как ключ партицирования. Apache Kafka гарантирует, что все сообщения с одинаковым ключом будут будут распределены в одни и те же партиции. По умолчанию false

### Сервис kafka_consumer
```
"kafka_consumer_options":{
            "brokers":"metrics01.example.org:9092",
            "topic":"metrics",
            "group":"metrics-reader",
            "base64_encode_body":false,
            "options":[
                "auto.offset.reset=earliest", 
                "enable.auto.commit=true", 
                "session.timeout.ms=6000"
            ]
         }, 
```
1. brokers: Список брокеров кафки в кластере-источнике. Опция обязательна
2. topic: из какого топика забирать сообщения. Опция обязательна
3. group: идентификатор группы потребителей в исходном кластере. При обрыве связи позволяет начать считывание сообщений с того места, с которого произошел обрыв, без потери сообщений (если активна опция enable.auto.commit=true). Опция обязательна
4. base64_encode_body: тело сообщения будет кодировано перед отправкой методом base64. По умолчанию false
5. options: настройки читателя из кафки. В программе используется библиотека librdkafka, список опций находится: [https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md](https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md). По умолчанию []

 
## Настройка записи в кластер(сервер) кафки
```
         "kafka_producer":{
            "brokers":"10.0.0.1:9092, 10.0.0.2:9092, 10.0.0.3:9092",
            "set_headers":{"from-server":"metrics01", "tag":"demo"},
            "topic":"metrics",
            "options":[
               "compression.codec=snappy",
               "socket.keepalive.enable=true",
               "queued.max.messages.kbytes=1048576",
               "retries=100000"
            ],
            "queue_buffer_len":10000
         }
```
У каждого сервиса индивидуальные настройки продюсера кафки, настройки продюсера описываются по отдельности в каждом активном сервисе.
1. brokers: Список брокеров кафки в кластере-назначении. Опция обязательна
2. set_headers: список заголовков и их значений, который будут добавлены в сообщение. Вне зависимости от протокола, они будут добавлены в служебный заголовок отправляемого в кафку сообщения. По умолчанию {}
3. topic: Топик, в который будут отправляться сообщения. Опция обязательна
4. options: настройки писателя в кафку. В программе используется библиотека librdkafka, список опций находится: [https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md](https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md)
5. queue_buffer_len: размер внутреннего буфера сообщений. По умолчанию 10000


### Сервис статистики
```
   "stat":{
      "listen":"0.0.0.0:8090",
      "json_path":"/stat"
      "metrics_path":"/metrics"
   }
```
1. listen: задаёт пару ip:port для прослушивания трафика. Опция обязательна
2. json_path: Путь, по которому будeт приниматься запросы на получение статистики методом GET. По умолчанию "/stat"
3. metrics_path: Путь для экспортирования метрик prometheus. По умолчанию "/metrics"

##### Примечание:
1. Пример получаемой статистики:
```
{
  "stats": [
    {
      "name": "http-stream-json",
      "output": {
        "producer_name": "rdkafka#producer-1",
        "total_processed_messages": 0,
        "current_ops": 0,
        "delivered_errors_count": 0,
        "delivered_messages_count": 0,
        "delivered_messages_bytes": 0,
        "messages_count_in_queue": 0,
        "messages_bytes_in_queue": 0,
        "max_queue_messages_count": 100000,
        "max_queue_messages_bytes": 1073741824
      }
    },
    {
      "name": "http-stream-test",
      "output": {
        "producer_name": "rdkafka#producer-2",
        "total_processed_messages": 0,
        "current_ops": 0,
        "delivered_errors_count": 0,
        "delivered_messages_count": 0,
        "delivered_messages_bytes": 0,
        "messages_count_in_queue": 0,
        "messages_bytes_in_queue": 0,
        "max_queue_messages_count": 100000,
        "max_queue_messages_bytes": 1073741824
      }
    }
  ]
}
```

2. Список метрик prometheus
```
kafka_bridge_total_processed_messages
kafka_bridge_messages_count_in_queue
kafka_bridge_messages_bytes_in_queue
kafka_bridge_delivered_errors_count
kafka_bridge_delivered_messages_count
kafka_bridge_delivered_messages_bytes
```


