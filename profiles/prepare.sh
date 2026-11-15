#!/bin/bash

echo '[{"id":"test1","type":"gauge","value":123.45},{"id":"test2","type":"counter","delta":1}]' > body.json
gzip -k body.json

hey -z 60s -c 20 -m POST \
  -H 'Content-Encoding: gzip' -H 'Content-Type: application/json' \
  -D body.json.gz \
  http://localhost:8080/updates/


#  -seconds=30 имеет смысл только для CPU

go tool pprof http://localhost:6060/debug/pprof/heap

go tool pprof -top -total_delay <file>