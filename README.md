# 2services

### Задача: Необходимо написать два сервиса, общающихся между собой по GRPC/Protobuf:

##### Сервис Cache:

- Cчитывает конфиг из файла config.yml (см ниже)
- Реализует GRPC метод GetRandomDataStream() без параметров, возвращающий поток из string
- Получив запрос  он NumberOfRequests(из конфига) раз в параллельных горутинах вызывает случайный URL из набора URLs(из конфига)
- Полученные ответы по мере их поступления отдаются через поток и кэшируются в БД Redis с временем жизни = случайным числом между MinTimeout(из конфига) и MaxTimeout(из конфига)
- При каждом подзапросе к URL необходимо сначала проверять наличие данных в БД Redis, и если они там есть, то отдавать их оттуда, а не через запрос по URL - т.е. выполнять фунцию “кэша”
- При этом если в БД Redis записей нет или они просрочены, то необходимо обращаться напрямую по URL.
- **ВАЖНО:** Сервис Cache будет запущен во множестве экземпляров на разных серверах. Поэтому другие параллельно выполняющиеся горутины в данном процессе а также в других процессах / на других серверах также не должны лезть в URL - они должны дожидаться окончания выполнения запроса первой горутиной и получить данные от нее или через Redis. Т.е. одновременно никогда от всех сервисов Cache не должно быть повторных обращений по одинаковым URL

Содержимое файла config.yml:
```
URLs:
- https://golang.org
- https://www.google.com
- https://www.bbc.co.uk
- https://www.github.com
- https://www.gitlab.com
- https://www.twitter.com
- https://www.facebook.com
MinTimeout: 10
MaxTimeout: 100
NumberOfRequests: 3
```

##### Сервис Consumer:

Создает 1000 горутин и в каждой из них делает запросы ко второму сервису. Результат можно никуда не выводить.

### Requirements:

- Go 1.13+ (project uses Go Modules)
- Docker + Docker Compose  - if you need to run it as docker containers

### Starting services 
1. Clone repository, `cd` to repository directory
1. Exec `./docker-build.sh` script. This will build go services and build docker images fro them
1. Run `docker-compose up`

Alternatively, you can start services from command line:
1. build services:<br>
`go build -o bin/cache github.com/artsv79/2services/cmd/cache`<br>
`go build -o bin/consumer github.com/artsv79/2services/cmd/consumer`

1. Edit config.yaml config file, if you need
1. Make sure you have Redis service running (e.g. its address "localhost:6379")
1. Run `./bin/cache -redis=localhost:6379`
1. Run `./bin/consumer -cache=localhost:9090`
<br> Please note, cache service has listening port number hardcoded to 9090. 
