# GYMLOG-GO
An application for tracking workouts written in Go. The application is built for learning purposes only. The application provides REST APIs for basic user actions and managing the sets of your workout. It uses PostgreSQL as a database and the authentication is done with JWT.

### Starting the application
1. Start a PostgreSQL-database container with the following command: ```docker run -it -p 5432:5432 --name=postgres_go -e POSTGRES_PASSWORD=password -d postgres```
2. Start the app by executing script "scripts/deploy.sh"

### Testing the application
1. Run the tests by executing script "scripts/test.sh"

### TODO:
- CSRF-protection (httpOnly cookie etc.)
- improve logging