# Проект «Асинхронные заказы и платежи»

## Общая структура и идея архитектуры

Система разнесена на несколько микросервисов, каждый со своей зоной ответственности:

**Gateway Service**
* Отвечает только за роутинг HTTP‑запросов
* – /orders* отправляются в сервис заказов, /accounts* – в сервис платежей
* – Имеет эндпоинт /health для проверки готовности

**Orders Service**
* Создаёт заказы, возвращает список заказов и статус отдельного заказа
* При создании заказа формирует событие «запрос на оплату» и кладёт его в таблицу outbox в одной транзакции
* Воркер отправляет события в Kafka. Второй воркер принимает результаты оплаты и обновляет статус заказа

**Payments Service**
* Создаёт счёт (один на пользователя), пополняет баланс, возвращает текущий баланс
* Потребляет события оплаты из Kafka, пытается списать деньги атомарно (Transactional Inbox + уникальные транзакции), записывает результат в outbox
* Отправляет событие результата обратно в Kafka


## Frontend

В проекте реализован простой frontend в виде статического web-интерфейса, предназначенного для ручной проверки работы системы.

* собран как отдельный сервис на базе **Nginx**
* проксирует все запросы вида `/api/*` в **Gateway Service**
* использует HTTP-запросы к API 

После запуска проекта интерфейс доступен по адресу: 
```html
http://localhost:8083
```



## User Flow

1. Клиент вызывает POST /accounts – создаётся счёт пользователя
2. Клиент может пополнить баланс через POST /accounts/topup
3. Клиент вызывает POST /orders. Gateway проксирует запрос в Orders; сервис создаёт заказ и пишет событие в outbox
4. Отдельный воркер Orders публикует событие в Kafka
5. Payments Service потребляет событие, проверяет, не дублируется ли оно (inbox), пытается списать деньги в транзакции (transactions), пишет результат в outbox
6. Воркер Payments публикует результат в Kafka
7. Orders Service обновляет статус заказа. Пользователь может просмотреть заказ (GET /orders/{order_id}) и увидеть статус NEW, FINISHED или FAILED


## API Gateway

Swagger‑документация доступна по адресу http://localhost:8080/swagger/index.html или в файле api/swagger.yaml

## Эндпоинты 

POST /accounts – создать счёт: { "user_id": UUID, "balance": number >= 0 }

POST /accounts/topup – пополнить счёт: { "user_id": UUID, "amount": number > 0 }

GET /accounts/{user_id} – получить баланс

POST /orders – создать заказ: { "user_id": UUID, "amount": number > 0, "description": string }. Возвращает order_id и статус NEW

GET /orders?user_id=… – получить список заказов пользователя

GET /orders/{order_id} – получить заказ

## Запуск

```bash
docker compose up -d --build
```


Система поднимет все сервисы, создаст топики Kafka и применит миграции. Проверить готовность можно по /health:
```bash
curl http://localhost:8080/health   # gateway
curl http://localhost:8081/health   # orders
curl http://localhost:8082/health   # payments
```

## Алгоритм списания и гарантии доставки

Система использует паттерны Transactional Outbox/Inbox для событий и идемпотентную логику:

Outbox в Orders: заказ и запись в outbox создаются в одной транзакции, исключая потерю событий.

Inbox в Payments: перед обработкой события запись message_id сохраняется в таблицу inbox; если такое сообщение уже есть, обработка не повторяется.

Transactions: запись в transactions с ключом order_id гарантирует, что деньги спишутся только один раз.

Обновление статуса заказа в Orders происходит условно (where status = 'NEW'), поэтому повторные результаты не изменят состояние.

## Примеры запросов (Bash)

```bash
# создание счета пользователя
USER_ID=$(uuidgen)
curl -s -X POST http://localhost:8080/accounts \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"'$USER_ID'","balance":500}'
```
```bash
# пополнение
curl -s -X POST http://localhost:8080/accounts/topup \
-H 'Content-Type: application/json' \
-d '{"user_id":"'$USER_ID'","amount":300}'
```
```bash
# создание заказа
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
-H 'Content-Type: application/json' \
-d '{"user_id":"'$USER_ID'","amount":200,"description":"книга"}' | grep -oE '[0-9a-f-]{36}')
```
```bash
# проверка
sleep 3
curl -s http://localhost:8080/orders/$ORDER_ID
curl -s http://localhost:8080/accounts/$USER_ID
```


## Форматы ответов

Успех:
```json
{
  "data": {
    "order_id": "6f4b0a0b-5a82-4d3a-8e9d-2f04135fbb38",
    "status": "NEW"
  }
}
```

Ошибка:
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "user_id is required"
  }
}

```

