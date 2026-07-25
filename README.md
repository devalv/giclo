# giclo

GitHub Liked repos cloner

## Установка и конфигурация

1. Загрузите соответствующую версию из раздела релизы
2. Скопируйте исполняемый файл в /usr/local/bin (или иной каталог доступный waybar на запуск) или установите deb-пакет
3. Создайте файл-конфигурации по инструкции описанной ниже
4. Проверьте запуск командой giclo -config /home/user/.config/giclo.yml

## Содержимое конфигурационного файла приложения (config.yml)

```
debug: true
gh_user: "github-username-which-liked-repos-to-download"
dir: "path-where-to-download-repos"
token: "github token"
```

### Параметры по умолчанию:

```
dir: "."
debug: false
```

## Сборка deb-пакета
По релизному тегу настроена автосборка и публикация в артефакты релиза на github.

### Локальная проверка сборки
make build-deb VERSION=0.3.2
