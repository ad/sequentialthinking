#!/bin/bash

echo "Сборка Sequential Thinking MCP Server на Go..."

# Проверяем наличие Go
if ! command -v go &> /dev/null; then
    echo "Go не установлен. Пожалуйста, установите Go 1.21 или новее."
    exit 1
fi

echo "✅ Внешние зависимости не требуются - используется только стандартная библиотека Go!"

# Собираем приложение
echo "Сборка приложения..."
go build -o sequentialthinking-server main.go

if [ $? -eq 0 ]; then
    echo "✅ Сборка завершена успешно!"
    echo "Исполняемый файл: ./sequentialthinking-server"
    echo "Размер: $(ls -lh sequentialthinking-server | awk '{print $5}')"
    echo ""
    echo "Для запуска выполните:"
    echo "  ./sequentialthinking-server"
else
    echo "❌ Ошибка при сборке"
    exit 1
fi
