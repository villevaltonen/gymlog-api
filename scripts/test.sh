# set env variables
export DB_NAME=postgres
export DB_USERNAME=postgres
export DB_PASSWORD=password
export DB_HOST=localhost
export JWT_KEY=my_secret_key

# run db container
docker stop postgres_test && docker rm postgres_test
docker run -it -p 5432:5432 --name=postgres_test -e POSTGRES_PASSWORD=password -d postgres
sleep 1

# build and run the app
go test -v ./app

# stop and remove the db container
docker stop postgres_test && docker rm postgres_test