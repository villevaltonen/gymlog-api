# set env variables
export DB_NAME=postgres
export DB_USERNAME=postgres
export DB_PASSWORD=password
export DB_HOST=localhost
export JWT_KEY=my_secret_key

# run db container
docker stop postgres_go && docker rm postgres_go
docker run -it -p 5432:5432 --name=postgres_go -e POSTGRES_PASSWORD=password -d postgres

# build and run the app
go build -o ./bin/gymlog-api
./bin/gymlog-api

# remove db container
docker stop postgres_api && docker rm postgres_go