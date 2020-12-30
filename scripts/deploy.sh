# set env variables
export APP_DB_NAME=postgres
export APP_DB_USERNAME=postgres
export APP_DB_PASSWORD=password

# run db container
docker run -it -p 5432:5432 --name=postgres_go -e POSTGRES_PASSWORD=password -d postgres

# build and run the app
go build -o ./bin/gymlog-go
./bin/gymlog-go