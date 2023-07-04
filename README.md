# go-grpc-server-boilerplate
A MongoDB gRPC Microservice Boilerplate written in Go.

## Getting Started

### Local Development
1. Get the Backend API from Github and configure the conf.json file:
   ```bash
   $ git clone https://github.com/JECSand/go-grpc-server-boilerplate
   $ cd go-grpc-server-boilerplate

2. Make SSL Certifications using provided script:
   ```bash
    $ make cert

3. configure boilerplate *If not using docker-compose:
    ```bash
   $ cp conf.json.example conf.json

4. run docker compose:
   ```bash
    $ docker-compose up -d
   
5. API Service will be reachable at: grpc://localhost:5555 with TLS enabled.