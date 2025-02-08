#!/bin/bash
          # Проверяем существование папки ./cmd/
          if [ ! -d "./cmd" ]; then
            echo "Error: Directory ./cmd does not exist."
            exit 1
          fi

          # Используем find для получения только директорий в ./cmd/, преобразуем результат в строку через пробелы
          dirs=$(find ./cmd -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)

          # Проверяем, что список не пустой
          if [ -z "$dirs" ]; then
            echo "Error: No directories found in ./cmd/"
            exit 1
          fi

          # Удаляем символы возврата каретки (\r) и преобразуем в строку через пробелы
          dirs=$(echo "$dirs" | tr '\r' ' ' | tr '\n' ' ')

          # Убираем лишние пробелы в конце строки
          dirs=$(echo "$dirs" | sed 's/[[:space:]]*$//')
          dirs=$(echo "$dirs" | jq -R 'split(" ")')
          dirs=$(echo "$dirs" | tr '\r' ' ' | tr '\n' ' ')
          dirs=$(echo "$dirs" | sed 's/[[:space:]]*$//')
          # Записываем результат в переменную окружения
          echo "application=$dirs"