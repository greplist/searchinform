# searchinform


### Task
Необходимо написать сервер, который по IP подключившегося клиента возвращал название
страны клиента. Сервер получать данные должен с помощью одного из следующих сервисов:
http://freegeoip.net/, http://geoip.nekudo.com (список можно дополнять)
Данные по IP адресам должны кэшироваться в локальной БД, должен быть установлен порог
«валидности» IP адреса в кэше.
Необходимо учесть, что провайдеры имеют ограничение на количество запросов, и при выборе
внешнего провайдера нужно проверять сколько запросов к нему было в течении 1 минуты
и переключаться на другой
Настройки параметров кэша и список провайдеров необходимо хранить/задавать в
конфигурационном файле.

### Start with Project
Clone git repository:
    cd $GOPATH
    git clone https://github.com/greplist/searchinform.git
    cd searchinform/

### Run tests:
    make tests

### Build and Run server
For run server, you may run:

    make

Now server is running with default config, is listenning port 8080, so you may test it:

    curl 127.0.0.1:8080/api/country?host=google.com
    curl 127.0.0.1:8080/api/country?host=google.com
    curl 127.0.0.1:8080/api/country?host=192.140.253.113

After 5 minutes, you may run:

    curl 127.0.0.1:8080/api/country?host=google.com

And server returns country for this host from real server, not from cache (cache TTL test)
