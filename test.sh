#!/bin/bash

echo "Тестирование Sequential Thinking MCP Server на Go..."

# Проверяем наличие собранного сервера
if [ ! -f "./sequentialthinking-server" ]; then
    echo "Сервер не найден. Запуск сборки..."
    make build-local
fi

# Запускаем unit тесты
echo "Запуск unit тестов..."
go test -v

# Создаем тестовые MCP запросы
echo ""
echo "Создание тестовых MCP запросов..."
cat > test_input.jsonrpc << 'EOF'
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}
{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "sequentialthinking", "arguments": {"thought": "Я начинаю анализировать эту проблему. Нужно разбить её на логические шаги.", "thoughtNumber": 1, "totalThoughts": 3, "nextThoughtNeeded": true}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "sequentialthinking", "arguments": {"thought": "Теперь я понимаю основные аспекты проблемы. Пересматриваю свои первоначальные предположения.", "thoughtNumber": 2, "totalThoughts": 4, "nextThoughtNeeded": true, "isRevision": true, "revisesThought": 1}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "sequentialthinking", "arguments": {"thought": "Финальный вывод основан на переосмыслении проблемы.", "thoughtNumber": 3, "totalThoughts": 3, "nextThoughtNeeded": false}}}
EOF

echo ""
echo "Запуск сервера с тестовыми MCP данными..."
echo "Нажмите Ctrl+C для остановки"
echo ""

cat test_input.jsonrpc | ./sequentialthinking-server -transport stdio

# Очистка
rm -f test_input.jsonrpc

echo ""
echo "Тестирование завершено."
