# Go-Freight: Inventory Service

Микросервис управления складскими запасами. Спроектирован по принципам Clean Architecture

## 🛠 Стек технологий

- **Язык:** Go 1.26
- **База данных:** PostgreSQL 15
- **Роутинг:** chi
- **Работа с БД:** sqlc (генерация типобезопасного кода), goose (миграции)
- **Конфигурация:** cleanenv
- **Логирование:** slog (структурированные логи с RequestID)
- **API Документация:** Swagger / OpenAPI
- **Тестирование:** testify, mockery
- **Инфраструктура:** Docker, Docker Compose, Taskfile

## 🏗 Архитектура

- **Clean Architecture:** Четкое разделение на слои `transport` (HTTP/REST), `service` (бизнес-логика) и `repository` (работа с БД).

## 📡 REST API

Документация Swagger UI доступна по адресу `/swagger/index.html`.

### Склады (Warehouses)

| Метод    | Эндпоинт                           | Описание                            |
| :------- | :--------------------------------- | :---------------------------------- |
| `POST`   | `/api/v1/warehouses`               | Регистрация нового склада           |
| `GET`    | `/api/v1/warehouses`               | Список всех складов                 |
| `GET`    | `/api/v1/warehouses/{id}`          | Детали конкретного склада           |
| `PUT`    | `/api/v1/warehouses/{id}`          | Обновление данных склада            |
| `DELETE` | `/api/v1/warehouses/{id}`          | Удаление склада                     |
| `GET`    | `/api/v1/warehouses/{id}/products` | Список товаров на конкретном складе |

### Товары и остатки (Products)

| Метод    | Эндпоинт                                   | Описание                                      |
| :------- | :----------------------------------------- | :-------------------------------------------- |
| `POST`   | `/api/v1/products`                         | Создание товара                               |
| `GET`    | `/api/v1/products/{id}`                    | Информация о товаре и его остатках            |
| `DELETE` | `/api/v1/products/{id}`                    | Удаление товара                               |
| `PATCH`  | `/api/v1/products/{id}/adjust`             | Корректировка остатка (+ приход / - списание) |
| `PATCH`  | `/api/v1/products/{id}/reserve`            | Резервирование товара под заказ               |
| `PATCH`  | `/api/v1/products/{id}/release`            | Окончательное списание после отгрузки         |
| `PATCH`  | `/api/v1/products/{id}/cancel-reservation` | Отмена брони (возврат в доступный остаток)    |
