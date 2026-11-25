Добавьте в пакет `cmd/shortener` (для трека «Сервис сокращения URL») или `cmd/server` и `cmd/agent` (для трека «Сервис сбора метрик и алертинга») глобальные переменные:

- `var buildVersion string`
- `var buildDate string`
- `var buildCommit string`

При старте приложения выводите в `stdout` сообщение в следующем формате:

Build version: `<buildVersion>` (или `N/A` при отсутствии значения)
Build date: `<buildDate>` (или `N/A` при отсутствии значения)
Build commit: `<buildCommit>` (или `N/A` при отсутствии значения) 