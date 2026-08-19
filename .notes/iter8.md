### Добавьте поддержку gzip в код сервера и агента. Научите:

- Агента передавать данные в формате `gzip`.
- Сервер опционально принимать запросы в сжатом формате (при наличии соответствующего HTTP-заголовка `Content-Encoding`).
- Отдавать сжатый ответ клиенту, который поддерживает обработку сжатых ответов (с HTTP-заголовком `Accept-Encoding`).

Функция сжатия должна работать для контента с типами `application/json` и `text/html`.

Вспомните middleware из урока про HTTP-сервер — это может вам помочь.

```bash
# Запрос с gzip-телом (агент шлёт сжатые данные)
# Чем полезен: Без "Content-Encoding: gzip" запрос упадет
echo '{"type":"gauge", "id":"cpu_usage","value":42}' | gzip > /tmp/body.gz
curl -v -X POST http://localhost:8080/update \
  -H "Content-Type: application/json" \
  -H "Content-Encoding: gzip" \
  --data-binary @/tmp/body.gz


# То же самое, но deflate
echo '{"type":"gauge","id":"cpu_usage","value":75}' | python3 -c "
import sys, zlib
data = sys.stdin.buffer.read()
co = zlib.compressobj(9, zlib.DEFLATED, -zlib.MAX_WBITS)
sys.stdout.buffer.write(co.compress(data) + co.flush())" > /tmp/body.deflate

curl -v -X POST http://localhost:8080/update \
  -H "Content-Type: application/json" \
  -H "Content-Encoding: deflate" \
  --data-binary @/tmp/body.deflate


# Ответ должен прийти сжатым (CompressMiddlewar)
# Чем полезен: смотрим в Content-Encoding в ответе
curl -v -H "Accept-Encoding: gzip" http://localhost:8080/


# Ответ без Accept-Encoding — как есть, без Content-Encoding
curl -v http://localhost:8080/


# Проверка Content-Length в ответе (должен совпадать с фактическим размером)
curl -s -H "Accept-Encoding: gzip" -o /tmp/resp.gz -D /tmp/headers.txt http://localhost:8080/
cat /tmp/headers.txt | grep -iE "content-(encoding|length)"
```
