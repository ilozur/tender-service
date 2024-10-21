# Avito Tender Service
Реализация серверной части сервиса для менеджмента системы тендеров. Выполнен для Avito. Подробнее про условия проекта можно почитать [здесь](task/README.md)

***
## Введение в проект

### Функциональность
Был реализован весь функционал предоставленный в файле [openapi.yml](task/openapi.yml). При работе над проектом отталкивался именно от примеров и заданий в этом файле

### Архитектура проекта
Была произведена попытка сделать наиболее чистую архитектуру
```
tender_service —
    — cmd
        — main
    — internal
        — config
        — handlers
            — bids
            — ping
            — tenders
        — lib
            — response
            — time_converter
        — storage
            — models
```
### Структура роутов
```
   /api
       /tenders
               /                                    — GET      — Получение списка тендеров
               /my                                  — GET      — Получение списка ваших тендеров
               /new                                 — POST     — Создание нового тендера
               /{tenderId}/status                   — GET      — Получение текущего статуса тендера
               /{tenderId}/status                   — PUT      — Изменение статуса тендера
               /{tenderId}/edit                     — PATCH    — Редактирование тендера
               /{tenderId}/rollback/{version}       — PUT      — Откат версии тендера
               
       /bids
               /my                                  — GET      — Получение списка ваших предложений
               /new                                 — POST     — Создание нового предложения
               /{bidId}/status                      — GET      — Получение текущего статуса предложения
               /{bidId}/status                      — PUT      — Изменение статуса предложения
               /{bidId}/edit                        — PATCH    — Редактирование параметров предложения
               /{bidId}/submit_decision             — PUT      — Отправка решения по предложению
               /{bidId}/feedback                    — PUT      — Отправка отзыва по предложению
               /{bidId}/rollback/{version}          — PUT      — Откат версии предложения
               /{tenderId}/list                     — GET      — Получение списка предложений для тендера 
               /{tenderId}/reviews                  — GET      — Просмотр отзывов на прошлые предложения
               
       /ping                                        — GET      — Проверка доступности сервера
```

### Таблицы в БД
```
   employee                      — Таблица с сотрудниками / пользователями
      id
      username
      first_name
      last_name
      created_at
      updated_at
      deleted_at
      
   organization                  — Таблица с организациями
      id
      name
      description
      type
      created_at
      updated_at
      deleted_at
   
   organization_responsible      — Таблица-связка сотрудника с организацией
      id
      organization_id
      user_id
   
   tenders                       — Таблица с тендерами
      id
      name
      description
      service_type
      status
      employee_username
      organization_id
      version
      created_at
      updated_at
      deleted_at
   
   tenders_versions              — Таблица с версиями тендеров
      id
      tender_id
      name
      description
      service_type
      status
      employee_username
      organization_id
      version
      created_at
      updated_at
      deleted_at
   
   bids                          — Таблица с предложениями
      id
      tender_id
      name
      description
      author_type
      status
      employee_username
      organization_id
      version
      created_at
      updated_at
      deleted_at
   
   bid_versions                  — Таблица с версиями предложений
      id
      tender_id
      bid_id
      name
      description
      author_type
      status
      employee_username
      organization_id
      version
      created_at
      updated_at
      deleted_at
   
   bid_feedbacks                 — Таблица с отзывами на предложения
      id
      bid_id
      feedback
      employee_username
      organization_id
      created_at
      updated_at
      deleted_at
```
### Использованные библиотеки
   * `chi` — Для работы с роутами
   * `gorm` — Для упрощения взаимодействия с БД
   * `pgx` — Для работы с PostgreSQL
   * `google/uuid` — Для генерации уникальных идентификаторов 
   * Множество стандартных библиотек GoLang:
     * `net/http`
     * `log/slog`
     * и т.д.
***
## Запуск проекта

### Необходимое ПО для запуска
   * GoLang 1.22.0+
   * PostgreSQL 14+
     * База данных должна так же быть заранее создана
   * Docker & Docker Compose

### Инструкция по запуску проекта
1. **Клонируйте репозиторий и переходите в директорию проекта:**
    ```shell
    git clone https://github.com/ilozur/tender-service.git
    cd tender-service
    ```
2. **Создайте файл для перемен окружения `.env` в корне проекта, заполнив его по ниже указанному образцу:**
   ```
   POSTGRES_HOST={localhost или host.docker.internal}
   POSTGRES_PORT={порт}
   POSTGRES_DATABASE={имя базы данных}
   POSTGRES_USERNAME={имя пользователя}
   POSTGRES_PASSWORD={пароль пользователя}
   ```
3. **Запустите сервис с помощью Docker Compose:**
    ```shell
    docker compose --env-file ./.env up
    ```
   
Поздравляю! Если всё сделано корректно, то теперь вы можете лицезреть работу сервиса по URL: http://localhost:8080/api
***
## Примечание к проекту
База данных изначально не содержит ни одной записи ни в одной из таблиц — перед началом работы крайне рекомендуется добавить пару записей в таблицы `employee`, `organization`, `organization_responsible`
