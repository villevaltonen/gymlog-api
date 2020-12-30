# GYMLOG-GO
An application for tracking workouts written in Go. The application is built for learning purposes only. The application provides REST APIs for basic user actions and managing the sets of your workout. It uses PostgreSQL as a database and the authentication is done with JWT.

### Running the application or tests
1. Start a PostgreSQL-database container with the following command: ```docker run -it -p 5432:5432 --name=postgres_go -e POSTGRES_PASSWORD=password -d postgres```
2. Initialize environment variables for connection: ```export APP_DB_USERNAME=postgres && export APP_DB_PASSWORD=password && export APP_DB_NAME=postgres```
3. Run the tests with ```go test -v``` or the app with ```go build && ./gymlog-go```

### TODO:
- pick user id from jwt and add to sql queries to ensure authorization
- REST APIs for user mgmt (create user etc.)