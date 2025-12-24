# About Mock Gateway
↩️ [Return to README](../../README.md)

This service uses a **mock payment gateway** to simulate interactions with an external payment provider. 
The mock gateway is designed to facilitate development and testing without relying on real payment processing systems.

It was done quickly in Python using FastAPI to focus on simulating various scenarios rather than implementing a full-fledged payment gateway.

The file mock-gateway/app.py exposes a POST endpoint /pay. It reads the JSON body and checks the mode field (defaults to happy). Each mode triggers different behavior:
- timeout — sleeps 10s, returns 504
- error — returns 500
- declined — returns 400
- latency — sleeps 1–3s, returns 200 with "approved"
- random — randomly picks one of the outcomes (may sleep for timeout) any other / default (happy) — returns 200 with "approved" 

To set the mode, send a JSON POST with the mode key. Example curl requests:



## Set mode to timeout
```curl
  curl -X POST http://localhost:8080/pay \
  -H "Content-Type: application/json" \
  -d '{"mode":"timeout"}'
```

## Set mode to error
```curl
  curl -X POST http://localhost:8080/pay \
  -H "Content-Type: application/json" \
  -d '{"mode":"error"}'
```

## Set mode to declined
```curl
curl -X POST http://localhost:8080/pay \
-H "Content-Type: application/json" \
-d '{"mode":"declined"}'
```

## Set mode to latency
```curl
curl -X POST http://localhost:8080/pay \
-H "Content-Type: application/json" \
-d '{"mode":"latency"}'
```

## Set mode to random
```curl
curl -X POST http://localhost:8080/pay \
-H "Content-Type: application/json" \
-d '{"mode":"random"}'
```

## Default (happy) — omit mode or set explicitly
```curl
curl -X POST http://localhost:8080/pay \
-H "Content-Type: application/json" \
-d '{}'
```