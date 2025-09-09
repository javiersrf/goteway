# Gateway API

## 🚀 Setup

### 🔑 Environment Variables
| Key              | Description                                                                 | Default         |
|------------------|-----------------------------------------------------------------------------|-----------------|
| `PROXYURL`       | Main URL for the upstream API services                                      | **required**    |
| `REQUEST_TIMEOUT`| Timeout (in seconds) for each request forwarded by the gateway              | `30`            |
| `REDISADDR`      | Redis server address (e.g. `localhost:6379`)                                | `localhost:6379`|
| `REDISPASSWORD`  | Redis password (leave empty if no password is set)                          | ``              |
| `REDISDB`        | Redis database index (integer)                                              | `0`             |
| `PORT`           | Port where the gateway will be exposed                                      | `8080`          |

---

## ▶️ Start Gateway

Run directly with Go:

```bash
go run main.go

```
Or build a binary and run:
```
go build -o gateway
./gateway
```


## 🌐 Access
The gateway will be available at:
```
http://localhost:8080
```
(unless overridden with the PORT environment variable).

## 🤝 Contributing

Contributions are welcome and appreciated!
If you’d like to improve this project, feel free to open an issue or submit a pull request.