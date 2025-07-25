# KMITL_FIET

TODO

- [x] Login
- [x] Register
- [ ] Update user
- [ ] Delete user
- [ ] Change password

## Setup Go

```sh
# go version
go version go1.24.5

# init
go mod init fiet

# install dependencies
go mod tidy

# install gin
# https://gin-gonic.com/docs/quickstart/
go get -u github.com/gin-gonic/gin@v1.10.1

# run
go run .
```

## Setup MSSQL

```sh
docker pull mcr.microsoft.com/mssql/server:2022-CU16-ubuntu-22.04

docker run -e "ACCEPT_EULA=Y" -e "MSSQL_SA_PASSWORD=Test1234" `
   -p 1433:1433 --name sql1 --hostname sql1 `
   -d mcr.microsoft.com/mssql/server:2022-CU16-ubuntu-22.04

docker run -e "ACCEPT_EULA=Y" -e "MSSQL_SA_PASSWORD=<password>" ^
    -p 1433:1433 ^
    -v <host directory>/data:/var/opt/mssql/data ^
    -v <host directory>/log:/var/opt/mssql/log ^
    -v <host directory>/secrets:/var/opt/mssql/secrets ^
    -d mcr.microsoft.com/mssql/server:2022-CU16-ubuntu-22.04

docker exec -it sql1 "bash"

# https://learn.microsoft.com/en-us/sql/linux/quickstart-install-connect-docker?view=sql-server-ver16&tabs=cli&pivots=cs1-bash
/opt/mssql-tools18/bin/sqlcmd -S localhost -U "sa" -P "Test1234" -C

/opt/mssql-tools18/bin/sqlcmd -S localhost -U "sa" -P "Test1234" -Q "SELECT name FROM sys.databases;" -C
```

## ENV

```sh
DB_USER="sa"
DB_PASSWORD="Test1234"
DB_SERVER="localhost"
DB_PORT="1433"
DB_DATABASE="test"
```

## Tree

```sh
├── main.go                    # Entry point
├── go.mod / go.sum           # Go modules
├── docker-compose.yaml       # Container orchestration
├── README.md
├── controller/               # Controllers (handlers)
├── database/                 # DB init/setup
├── middleware/               # Gin middleware
├── model/                    # Structs and DB models
├── router/                   # Route definitions
├── service/                  # Business logic / JWT svc
```

## Database Library

1. https://jmoiron.github.io/sqlx/

Database fiet

User permission

```sh
-- 1. Create login at server level
CREATE LOGIN fiet_user WITH PASSWORD = 'StrongP@ssw0rd';

-- 2. Switch to your application database
USE fiet;

-- 3. Create user in the database mapped to the login
CREATE USER fiet_user FOR LOGIN fiet_user;

-- 4. Grant permissions (basic CRUD access)
EXEC sp_addrolemember 'db_datareader', 'fiet_user'; -- SELECT
EXEC sp_addrolemember 'db_datawriter', 'fiet_user'; -- INSERT, UPDATE, DELETE

-- Optional: grant execute for stored procedures
-- GRANT EXECUTE TO fiet_user;
```

CREATE USERS

```sh
CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    uuid NVARCHAR(36) NOT NULL DEFAULT CONVERT(NVARCHAR(36), NEWID()) UNIQUE, -- public identifier
    name NVARCHAR(100),
    email NVARCHAR(100) NOT NULL UNIQUE,
    age INT CHECK (age >= 0 AND age <= 150),
    password_hash NVARCHAR(255) NOT NULL,
    created_at DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
    updated_at DATETIME2 NOT NULL DEFAULT SYSDATETIME()
);
```