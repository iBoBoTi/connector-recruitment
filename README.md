# Connector Service (Go + Buf + LocalStack)
- **gRPC** service with three methods:
    - `CreateConnector` (You are given static access tokens and the default channel name(which needs to be resolved to its ID). See [Bonus](#bonus) for OAuthV2)
    - `GetConnector`
    - `DeleteConnector`
- **Secrets Manager** integration (LocalStack).
- **Slack integration** to send messages using an already created connector.
- **Optional PostgreSQL** usage for tracking connector metadata.

---

## Architecture

```
  +-----------------+         +--------------------+
  | gRPC Client     | ----->  | Slack Connector    |
  | (e.g., grpcurl) |         | Service (Go + Buf) |
  +-----------------+         +--------------------+
        |                                 |
        | (AWS SDK)                       | (Slack API)
        v                                 v
  LocalStack (Secrets Manager)       Slack (Real or Mock)
        |
   +------------+
   | PostgreSQL |
   +------------+
```

# **Connector Service**

A Golang-based, clean-architecture application for managing Slack connectors. It stores static Slack tokens in AWS Secrets Manager (optionally via LocalStack) and persists connector metadata in PostgreSQL. You can also send notifications/messages via Slack channels.

---

## **Features**
### **1. Connector Management**

This service also exposes gRPC endpoints defined in connector.proto. Below are some key methods:

- **Create Connector** 
  **Request (Protobuf):**
  ```protobuf
      message CreateConnectorRequest {
      string workspace_id = 2;
      string tenant_id = 3;
      string default_channel_name = 4;
      string slack_token = 5;
   }
   ```
   **Response (Protobuf):**
   ```protobuf
   message CreateConnectorResponse {
      string id = 1;
      string workspace_id = 2;
      string tenant_id = 3;
      string default_channel_id = 4;
      string created_at = 5;
      string updated_at = 6;
   }
   ```
- **Get Connector** 
  **Request (Protobuf):**
  ```protobuf
   message GetConnectorRequest {
      string connector_id = 1;
   }
   ```
   **Response (Protobuf):**
   ```protobuf
   message GetConnectorResponse {
      string id = 1;
      string workspace_id = 2;
      string tenant_id = 3;
      string default_channel_id = 4;
      string created_at = 5;
      string updated_at = 6;
   }
   ```
- **Delete Connector** 
  **Request (Protobuf):**
  ```protobuf
   message DeleteConnectorRequest {
      string connector_id = 1;
   }
   ```
   **Response (Protobuf):**
   ```protobuf
   message DeleteConnectorResponse {
      bool success = 1;
   }
   ```

## **Quick Start: Local Development**

### **1. Clone the Repository**
```bash
   git clone https://github.com/iBoBoTi/connector-service.git
   cd connector-service
```
Generate protobuf stubs (if you haven’t yet):
```bash
   buf generate
```

### **2. Export Environment Variables**
Export environment variables with your own configuration bearing in mind the system comes with its own default configuration:
```bash
   export DB_HOST=db
   export DB_PORT=5432
   export DB_USER=aryon
   export DB_PASSWORD=aryon
   export DB_NAME=aryondb
   export AWS_REGION=us-east-1
   export AWS_ENDPOINT=http://localhost:4566
   export GRPC_PORT=50051
```

### **3. Build and Run the Application**
Use Docker Compose to build and run:
```bash
   make start-services
```
Spins up PostgreSQL (for connector metadata)
Spins up LocalStack (for Secrets Manager)
Builds and starts Slack Connector Service
At start up the database migration runs

### **4. Verify gRPC**
 Use grpcurl or any gRPC client to test endpoints, e.g.,
```bash
grpcurl -plaintext \
  -d '{"workspace_id":"WS123","tenant_id":"TNT123","default_send_channel_name":"general","slack_token":"valid-token"}' \
  localhost:50051 connector.v1.SlackConnectorService/CreateConnector
```

## **Automated Testing**
Run unit and integration tests:
```bash
`  go test -v -cover ./...
```
## **Technologies Used**
Programming Language: Golang

Frameworks/Libraries:

gRPC + Buf for Protobuf and code generation

AWS SDK for Secrets Manager (LocalStack in dev)

PostgreSQL driver + migrations

Containerization: Docker, Docker Compose

Testing: Go’s built-in testing framework, plus testify for assertions/mocks