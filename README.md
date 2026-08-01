# CareU

CareU is a local Go + MySQL prototype for personal healthcare review.

## Requirements

- Go
- MySQL or XAMPP MySQL

## Database

The server can create the `careu` database and tables automatically when the MySQL user has permission.

You can also import `schema.sql` manually in phpMyAdmin or MySQL.

Default local settings:

```powershell
$env:CAREU_DB_ADDR="127.0.0.1:3306"
$env:CAREU_DB_NAME="careu"
$env:CAREU_DB_USER="root"
$env:CAREU_DB_PASS=""
```

## Run

```powershell
cd C:\Projects\Lucky
go run .
```

Open:

```text
http://localhost:8080
```

Use the Register page first, then login/profile data will be stored in MySQL.
