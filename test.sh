#!/bin/bash

echo "Тестирование Sequential Thinking MCP Server на Go..."

# Проверяем наличие собранного сервера
if [ ! -f "./sequentialthinking-server" ]; then
    echo "Сервер не найден. Запуск сборки..."
    ./build.sh
fi

# Создаем тестовые запросы в отдельных строках
echo '{"id": "1", "method": "tools/list"}' > test_input.txt
echo '{"id": "2", "method": "tools/call", "params": {"name": "sequentialthinking", "arguments": {"thought": "Я начинаю анализировать эту проблему. Нужно разбить её на логические шаги.", "thoughtNumber": 1, "totalThoughts": 3, "nextThoughtNeeded": true}}}' >> test_input.txt
echo '{"id": "3", "method": "tools/call", "params": {"name": "sequentialthinking", "arguments": {"thought": "Теперь я понимаю основные аспекты проблемы. Пересматриваю свои первоначальные предположения.", "thoughtNumber": 2, "totalThoughts": 4, "nextThoughtNeeded": true, "isRevision": true, "revisesThought": 1}}}' >> test_input.txt

echo "Запуск сервера с тестовыми данными..."
echo "Нажмите Ctrl+C для остановки"

cat test_input.txt | ./sequentialthinking-server

# Очистка
rm -f test_input.txt

echo "Тестирование завершено."
