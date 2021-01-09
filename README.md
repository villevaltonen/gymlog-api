# GYMLOG-GO

An application for tracking workouts written in Go. The application is built for learning purposes only. The application provides REST APIs for basic user actions and managing the sets of your workout. It uses PostgreSQL as a database and the authentication is done with JWT.

### Starting the application

1. Start the app by executing script "scripts/deploy-compose.sh" (deploys the application and the database with Docker compose)

### Testing the application

1. Run the tests by executing script "scripts/test.sh" (the application runs locally, the database in a container)

### TODO

- CORS-origin to env-variable
- Clean up error messages to client
