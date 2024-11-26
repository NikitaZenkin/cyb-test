- Reverse DNS server -

Для локального запуска нужно иметь docker desktop

Чтобы поднять приложение, находясь в корневой директории проекта вызовете команду
> docker-compose --env-file config/.env up --build

При успешном запуске api будет доступно по адресу
- http://localhost:8081/api

Документация доступна по адресу
- http://localhost:8081/swagger/index.html
