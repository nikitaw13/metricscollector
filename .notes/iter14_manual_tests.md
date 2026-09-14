```bash
go build -o /tmp/server ./cmd/server && /tmp/server -k testkey
```

```bash
BODY='{"id":"t1","type":"gauge","value":1.5}'
HASH=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "testkey" -hex | awk '{print $2}')
```

#### Тест 1 — правильный хеш (ждём 200)

```bash
curl -i -X POST http://localhost:8080/update/ \
  -H "Content-Type: application/json" \
  -H "HashSHA256: $HASH" \
  -d "$BODY"
```

#### Тест 2 — мусорный хеш (ждём 400):

```bash
curl -i -X POST http://localhost:8080/update/ \
  -H "Content-Type: application/json" \
  -H "HashSHA256: deadbeef" \
  -d "$BODY"
```

#### Тест 3 — без заголовка вообще (ждём 400 при включённом ключе)

```bash
curl -i -X POST http://localhost:8080/update/ \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

#### Тест 4 — сервер без ключа 

перезапусти без `-k` — тот же запрос без заголовка → ждём 200