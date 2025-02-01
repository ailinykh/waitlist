# waitlist

A simple web service for using as a placeholder for your future awesome telegram bot projects

## Setup

```
go mod download
make .
```

## Roadmap

no milestones yet

### Features

- [x] handle POST request from telegram bot api
- [x] allow user to join waitlist
- [x] respond to healthcheck ping
- [ ] serve HTML page at the root `/` with waitlist table

### NFR

- [x] support `X-Telegram-Bot-Api-Secret-Token` to avoid security issues
- [x] save payload along with user data
- [ ] protect API with JWT-authorization