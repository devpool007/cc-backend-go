# README.md

# WebRTC Server

This project is a WebRTC server that facilitates real-time communication between users. It uses WebSocket connections to manage user interactions and room management.

## Project Structure

- `cmd/server/main.go`: Entry point of the application. Initializes the server and starts listening for incoming connections.
- `internal/server/server.go`: Contains the main server logic, including WebSocket setup and user interaction handling.
- `go.mod`: Defines the module's dependencies and specifies the module's name.
- `go.sum`: Contains checksums for the module's dependencies to ensure integrity.

## Getting Started

To run the server, navigate to the `cmd/server` directory and execute the following command:

```bash
go run main.go
```

## Dependencies

Make sure to run `go mod tidy` to install the necessary dependencies defined in `go.mod`.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.