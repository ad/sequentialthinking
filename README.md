# Sequential Thinking MCP Server

🧠 **Интеллектуальный MCP сервер** для пошагового анализа и решения сложных задач с поддержкой **двух режимов работы**: 
- **Stdio режим** для полной совместимости с MCP клиентами (VS Code, Claude Desktop)
- **HTTP+SSE режим** для веб-интерфейса и расширенной отладки

*Обновлено: 13 июня 2025*

## ⚡ Быстрый старт

### 1. Сборка проекта
```bash
# Клонирование репозитория
git clone <repository-url>
cd sequentialthinking

# Автоматическая сборка (рекомендуется)
./build.sh

# Или ручная сборка
go build -o sequentialthinking-server main.go
```

### 2. Запуск сервера

#### 📡 Для MCP клиентов (VS Code, Claude Desktop)
```bash
./sequentialthinking-server --stdio
```

#### 🌐 Для веб-разработки и отладки  
```bash
./sequentialthinking-server
# или с настройкой порта
PORT=3000 ./sequentialthinking-server
```

### 3. Тестирование
```bash
# Автоматическое тестирование
./test.sh

# Веб-интерфейс
open http://localhost:8080
```



## 🚀 Использование

### Интеграция с VS Code
Создайте файл `.vscode/mcp.json` в корне проекта:
```json
{
  "servers": {
    "sequentialthinking": {
      "type": "stdio", 
      "command": "/absolute/path/to/sequentialthinking-server",
      "args": ["--stdio"]
    }
  }
}
```

### Интеграция с Claude Desktop
Добавьте в `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "sequentialthinking": {
      "command": "/absolute/path/to/sequentialthinking-server",
      "args": ["--stdio"]
    }
  }
}
```

### Docker развертывание
```bash
# Сборка образа
docker build -t sequentialthinking .

# Запуск контейнера
docker run --rm -i sequentialthinking
```

## 🔍 Тестирование и отладка

### Автоматическое тестирование
```bash
# Запуск всех тестов
./test.sh

# Запуск Go unit тестов
go test -v

# Ручное тестирование stdio режима
echo '{"id": "1", "method": "tools/list"}' | ./sequentialthinking-server --stdio
```

### HTTP API тестирование
```bash
# Запуск сервера в фоне
PORT=8083 ./sequentialthinking-server &

# Список доступных инструментов
curl -X POST http://localhost:8083/mcp \
  -H "Content-Type: application/json" \
  -d '{"id": "1", "method": "tools/list"}'

# Вызов sequential thinking
curl -X POST http://localhost:8083/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "id": "2", 
    "method": "tools/call",
    "params": {
      "name": "sequentialthinking",
      "arguments": {
        "thought": "Анализирую эту проблему пошагово",
        "thoughtNumber": 1,
        "totalThoughts": 3,
        "nextThoughtNeeded": true
      }
    }
  }'

# Подключение к SSE потоку
curl -N http://localhost:8083/events
```

### Веб-интерфейс
Откройте `http://localhost:8080` для интерактивного тестирования с визуальным интерфейсом.

## 🧠 Инструмент Sequential Thinking

Предоставляет структурированный подход к решению сложных задач через пошаговое мышление.

### Основные параметры:
- **`thought`** *(string)*: Текущий шаг размышления  
- **`thoughtNumber`** *(integer)*: Номер текущей мысли (начиная с 1)
- **`totalThoughts`** *(integer)*: Оценочное общее количество шагов
- **`nextThoughtNeeded`** *(boolean)*: Требуется ли следующий шаг

### Дополнительные параметры:
- **`isRevision`** *(boolean)*: Пересматривает ли данная мысль предыдущую
- **`revisesThought`** *(integer)*: Номер пересматриваемой мысли
- **`branchFromThought`** *(integer)*: Точка ветвления для альтернативных подходов  
- **`branchId`** *(string)*: Идентификатор ветки
- **`needsMoreThoughts`** *(boolean)*: Индикатор необходимости дополнительных шагов

### Примеры использования:

#### Базовое последовательное мышление:
```json
{
  "thought": "Начинаю анализ алгоритма сортировки",
  "thoughtNumber": 1,
  "totalThoughts": 5, 
  "nextThoughtNeeded": true
}
```

#### Пересмотр предыдущего решения:
```json
{
  "thought": "Пересматриваю выбор структуры данных с учетом требований производительности",
  "thoughtNumber": 3,
  "totalThoughts": 6,
  "nextThoughtNeeded": true,
  "isRevision": true,
  "revisesThought": 2
}
```

## 🔧 Режимы работы и архитектура

### 📡 Stdio режим (MCP совместимость)
- **Назначение**: Интеграция с MCP клиентами (VS Code, Claude Desktop)
- **Протокол**: JSON-RPC через stdin/stdout
- **Запуск**: `./sequentialthinking-server --stdio`
- **Особенности**: Полная совместимость с MCP спецификацией 2025-03-26

### 🌐 HTTP+SSE режим (Веб-интерфейс)
- **Назначение**: Отладка, тестирование, веб-интеграция
- **Протокол**: HTTP API + Server-Sent Events
- **Запуск**: `./sequentialthinking-server` (по умолчанию порт 8080)
- **Эндпоинты**:
  - `GET /` - Веб-интерфейс для тестирования
  - `POST /mcp` - MCP запросы в JSON формате  
  - `GET /events` - SSE поток событий реального времени

### Преимущества dual-mode архитектуры:
✅ **Гибкость**: Один сервер для разных сценариев использования  
✅ **Отладка**: Веб-интерфейс для тестирования и мониторинга  
✅ **Совместимость**: Полная поддержка MCP стандарта  
✅ **Масштабируемость**: HTTP режим поддерживает множественные подключения



## 🛠️ Технические детали

### Системные требования
- **Go**: версия 1.24 или новее
- **ОС**: Linux, macOS, Windows  
- **Зависимости**: Только стандартная библиотека Go (без внешних пакетов)

### Структура проекта
```
sequentialthinking/
├── main.go              # Основной код сервера
├── main_test.go         # Unit тесты  
├── go.mod               # Go модуль
├── build.sh             # Скрипт автоматической сборки
├── test.sh              # Скрипт тестирования
├── Dockerfile           # Конфигурация Docker
├── .gitignore           # Правила игнорирования Git
├── README.md            # Документация
├── EXAMPLES_USAGE.md    # Примеры использования
└── VS_CODE_USAGE.md     # Руководство по VS Code
```

### Сборка и развертывание
```bash
# Клонирование и сборка
git clone <repository-url>
cd sequentialthinking
./build.sh

# Ручная сборка
go build -o sequentialthinking-server main.go

# Сборка для разных платформ
GOOS=linux GOARCH=amd64 go build -o sequentialthinking-linux main.go
GOOS=windows GOARCH=amd64 go build -o sequentialthinking.exe main.go
GOOS=darwin GOARCH=arm64 go build -o sequentialthinking-macos main.go
```

### Конфигурация
- **Порт HTTP сервера**: переменная окружения `PORT` (по умолчанию 8080)
- **Логирование**: все логи выводятся в stderr
- **Режим работы**: определяется наличием флага `--stdio`

---

**✨ Статус проекта**: Активная разработка  
**🔄 Последнее обновление**: 13 июня 2025  
**📝 Лицензия**: MIT License

---

## 📚 Дополнительные ресурсы

- 🌐 **[Model Context Protocol](https://modelcontextprotocol.io/)** - Официальная документация MCP
- 🐙 **[GitHub Copilot Chat](https://docs.github.com/en/copilot/github-copilot-chat)** - Документация по Copilot Chat

## 🤝 Поддержка

Если у вас возникли вопросы или проблемы:
1. Проверьте документацию в этом репозитории
2. Изучите логи сервера в stderr
3. Попробуйте веб-интерфейс для отладки
4. Создайте issue в репозитории проекта
