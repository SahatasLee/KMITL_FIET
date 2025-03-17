# KMITL_FIET

## Setup Go

```sh
# go version
go version go1.23.4 windows/amd64

# init
go mod init fiet

# install dependencies
go mod tidy

# install gin
# https://gin-gonic.com/docs/quickstart/
# require github.com/gin-gonic/gin v1.10.0
go get -u github.com/gin-gonic/gin@v1.10.0
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

/opt/mssql-tools18/bin/sqlcmd -S localhost -U "sa" -P "Test1234" -C
```