#!/bin/bash
# Проверяем существование папки ./cmd/
if [ ! -d "./cmd" ]; then
    echo "Error: Directory ./cmd does not exist."
    exit 1
fi

# Используем find для получения только директорий в ./cmd/
dirs=$(find ./cmd -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)
if [ -z "$dirs" ]; then
    echo "Error: No directories found in ./cmd/"
    exit 1
fi

# Записываем результат в переменную окружения
echo "dirs=$dirs" >> $TTT
echo $dirs