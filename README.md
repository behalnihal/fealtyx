# Student Management API with Ollama Integration

A RESTful API for managing student records with AI-powered summaries using Ollama.

## Features

- ✅ **Go Module**: Properly initialized with `go mod init`
- ✅ **REST API Endpoints**: Complete CRUD operations
- ✅ **Data Storage**: In-memory storage using slices
- ✅ **Ollama Integration**: AI-powered student summaries
- ✅ **Error Handling**: Comprehensive error handling
- ✅ **Input Validation**: Data validation for all inputs
- ✅ **Concurrency**: Thread-safe operations with mutex

## API Endpoints

### 1. Create a Student

```bash
POST /students
Content-Type: application/x-www-form-urlencoded

name=John Doe&age=20&email=john.doe@example.com
```

### 2. Get All Students

```bash
GET /students
```

### 3. Get Student by ID

```bash
GET /students/{id}
```

### 4. Update Student

```bash
PUT /students/{id}
Content-Type: application/x-www-form-urlencoded

name=John Smith&age=21&email=john.smith@example.com
```

### 5. Delete Student

```bash
DELETE /students/{id}
```

### 6. Generate Student Summary (with Ollama)

```bash
GET /students/{id}/summary
```

## Setup and Running

### Prerequisites

1. Go 1.23.2 or later
2. Ollama installed and running locally

### Install Ollama

1. Download from [https://ollama.ai](https://ollama.ai)
2. Install and start Ollama
3. Pull a model: `ollama pull llama2`

### Run the API

```bash
go run main.go
```

The server will start on `http://localhost:8000`

## Testing the API using Postman

Import these requests:

1. **POST** `http://localhost:8000/students`

   - Body: `x-www-form-urlencoded`
   - Key-value pairs: `name`, `age`, `email`

2. **GET** `http://localhost:8000/students`

3. **GET** `http://localhost:8000/students/1`

4. **PUT** `http://localhost:8000/students/1`

   - Body: `x-www-form-urlencoded`
   - Key-value pairs: `name`, `age`, `email`

5. **GET** `http://localhost:8000/students/1/summary`

6. **DELETE** `http://localhost:8000/students/1`

## Notes

- The API uses form data for POST/PUT requests
- All responses are in JSON format
- Student IDs are auto-generated (1, 2, 3, ...)
- Ollama must be running on `localhost:11434` for summary generation
- The default model is `llama3.2` - change it in the code if needed
