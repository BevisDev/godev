# GoDev

This module contains helper utilities and libraries to support developers in simplifying common tasks, improving code
reusability, and enhancing development efficiency

## Getting started

***Prerequisites***

- [Go 1.23.4](https://go.dev/doc/install) or higher

then install

```sh
go get github.com/BevisDev/godev@latest
```

## Dependencies

- [Configuration](#getting-viper)
- [Zap](#getting-zap)
- [Lumberjack](#getting-lumberjack)
- [Cron](#getting-cron)
- [Database](#getting-database)
- [Redis](#getting-redis)
- [Migration](#getting-migration)
- [Keycloak](#getting-keycloak)

Helper

- [UUID](https://github.com/google/uuid)
- [RabbitMQ](#getting-rabbitmq)
- [Decimal](https://github.com/shopspring/decimal)

### Technology stack

> **Note:**
>
> To switch to a difference environment, you need to set the environment variable
>
> On Windows:
>
> ```sh
> setx GO_PROFILE dev
> ```
>
> On Linux:
>
> ```sh
> export GO_PROFILE=dev
> ```

```sh
choco install make
```

On Linux using `apt` to install

```sh
sudo apt update
sudo apt install make
```

### Getting Viper

Document: [Viper](https://github.com/spf13/viper)

```sh
go get github.com/spf13/viper
```

### Getting Zap

Document: [Zap](https://github.com/uber-go/zap)

```sh
go get -u go.uber.org/zap
```

### Getting Lumberjack

Document: [Lumberjack](https://github.com/natefinch/lumberjack)

```sh
go get github.com/natefinch/lumberjack
```

### Getting Cron

Document: [Cron](https://github.com/robfig/cron)

```sh
go get github.com/robfig/cron/v3@v3.0.0
```

**Cron Expression Format**

A cron expression represents a set of times, using 6 space-separated fields.

#### Example:

```go
c := cron.New(cron.WithSeconds())
// second minute hour day month weekday
c.AddFunc("0 * * * * *", func () {
fmt.Println("Running every minute at the 0th second!")
})
c.Start()
```

| Field name   | Mandatory? | Allowed values  | Allowed special characters |
|--------------|------------|-----------------|----------------------------|
| Seconds      | Yes        | 0-59            | * / , -                    |
| Minutes      | Yes        | 0-59            | * / , -                    |
| Hours        | Yes        | 0-23            | * / , -                    |
| Day of month | Yes        | 1-31            | * / , - ?                  |
| Month        | Yes        | 1-12 or JAN-DEC | * / , -                    |
| Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?                  |

### Getting Database

***Install Driver***

```sh
go get github.com/denisenkom/go-mssqldb #MSSQL
go get github.com/lib/pq #Postgresql
go get github.com/godror/godror@latest #Oracle
```

- [SQL Server](https://github.com/denisenkom/go-mssqldb)
- [PostgreSQL](https://github.com/lib/pq)
- [Oracle](https://github.com/godror/godror)
- [Other Driver](https://go.dev/wiki/SQLDrivers)

To use map into struct easily:

```sh
go get github.com/jmoiron/sqlx
```

### Getting Redis

Document: [Redis](https://github.com/redis/go-redis)

```sh
go get github.com/redis/go-redis/v9
```

### Getting RabbitMQ

Document: [RabbitMQ](https://github.com/rabbitmq/amqp091-go)

```sh
go get github.com/rabbitmq/amqp091-go
```

- ack(multiple bool)
    - true: confirm all message sucessful
    - false: confirm one message sucessfuly

- nack(multiple, requeue bool)
    - multiple like ack

### Getting Migration

Document: [Goose](https://github.com/pressly/goose)

```sh
go get github.com/pressly/goose/v3/cmd/goose@latest
```

### Getting Keycloak

Document: [Gocloak](https://github.com/Nerzal/gocloak)

```sh
go get github.com/Nerzal/gocloak/v13
```