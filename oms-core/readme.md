# OMS Core service

## Database migration

Install goose, then use below command:

```bash
goose -dir migrations postgres "postgres://devuser:devpassword@localhost:15432/devdb?sslmode=disable" up
```