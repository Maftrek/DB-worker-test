Сервис слушает nats-streaming, куда отправляются запросы с сервиса лежащего в репозитории template, обрабатывает их и получает ответ от бд postgres (Protocol Buffers — протокол сериализации).
Так же сервис отвечает на запросы (примеры описаны ниже), используя mongo db.

---
Для работы на компьютере должен быть установлен nats-streaming-server, postgres и mongo db.

1. Nats server необходимо запустить командой nats-streaming-server.

2. Настройка mongo
* Подключится к экземпляру MongoDB (mongo)
* Создать пользователя admin с паролем 123
* Дать права (db.grantRolesToUser('admin', [{ role: 'root', db: 'admin' }]))
* Выйти (exit)
* Перезапустить mongo (mac OS - brew services restart mongodb-community@4.2)

3. Для postgres создать пользователя - CREATE USER role_1 WITH PASSWORD '1';

---

Порт приложения - :8120

#### Методы для обращения к базе данных mongo db:

**GET /api/news/{title} -  получить новость по названию (title)**

**POST /api/one_news/ - создать новость** 
body (json)
```json
{
	"title": "Name",
	"date": "2019-01-01"
}
```

**POST /api/many_news/ - создать несколько новостей** 
body (json)
```json
[
{
	"title": "Name",
	"date": "2019-01-01"
},
{
	"title": "Name new",
	"date": "2019-01-01"
}
]
```

**PUT /api/news/ - создать новость** 
body (json)
```json
{
	"title_old": "Name old",
	"title_new": "Name new"
}
```
---
Есть возможность развернуть докер образ, но в этом случае потребуется иметь nats-streaming-server и postgres не локально. Если условие выполняется то адреса nats-streaming-server и postgres необходимо прописать в config/config.toml

[NatsServer]
  Address = "АДРЕСС NATS"

[SQLDataBase]
  Server = "АДРЕСС POSTGRES"
  Database = "postgres"
  ApplicationName = ""
  MaxIdleConns = 20
  MaxOpenConns = 20
  ConnMaxLifetime = "5m"
  Port = 5432
  
[NoSQLDataBase]
  Server = АДРЕСС MONGO"
  Port = 27017
  Database = "test"
  
Так же потребуется создать Для postgres пользователя - CREATE USER role_1 WITH PASSWORD '1';
Для mongo db необходимо будет выплнить настройку, описанную выше

---

Для запуска докер образа необходимо выполнить в директории программы:

1) CGO_ENABLED=0 GOOS=linux go build -o main .
2) docker build -t tmp_db:0.0.1 .
3) docker run tmp_db:0.0.1
