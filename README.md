# GYMLOG-GO
An example application for tracking workouts written in Go. The application provides REST APIs for basic user actions and managing the sets of your workout. It uses PostgreSQL as a database and the authentication is done with JWT.

### Running the application or tests
1. Start a PostgreSQL-database container with the following command: ```docker run -it -p 5432:5432 --name=postgres_go -e POSTGRES_PASSWORD=password -d postgres```
2. Initialize environment variables for connection: ```export APP_DB_USERNAME=postgres && export APP_DB_PASSWORD=password && export APP_DB_NAME=postgres```
3. Run the tests with ```go test -v``` or the app with ```go build && ./gymlog-go```

### TODO:
- Script the testing from container setup to tearing it down
- Finish Docker compose
- JWT-authentication
- REST APIs for user mgmt