package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

var (
	students []Student
	mutex    sync.RWMutex
)

func handleStudents(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()
	
	// Convert students to JSON
	jsonData, err := json.Marshal(students)
	if err != nil {
		http.Error(w, "Error marshaling data", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func validateStudent(student Student) error {
	if student.Name == "" {
		return fmt.Errorf("name is required")
	}
	if student.Age <= 0 || student.Age > 150 {
		return fmt.Errorf("age must be between 1 and 150")
	}
	if student.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

func callOllamaAPI(student Student) (string, error) {
	prompt := fmt.Sprintf("Generate a brief, friendly summary of this student: Name: %s, Age: %d, Email: %s. Keep it under 100 words. Don't include any other text like 'Here is the summary' or 'Here is the student' or 'Here is the student summary'. Just the summary.", 
		student.Name, student.Age, student.Email)
	
	requestBody := OllamaRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Stream: false,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned status: %d", resp.StatusCode)
	}
	
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}
	
	return ollamaResp.Response, nil
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}


func main() {
	students = []Student{}
	api := http.NewServeMux()

	// introduction page

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the Student Management API\n"))
		w.Write([]byte("You can use the following endpoints to manage students\n"))
		w.Write([]byte("GET /students - Get all students\n"))
		w.Write([]byte("POST /students - Create a new student\n"))
		w.Write([]byte("PUT /students/{id} - Update a student\n"))
		w.Write([]byte("DELETE /students/{id} - Delete a student\n"))
		w.Write([]byte("GET /students/{id}/summary - Get a summary of a student\n"))
	})

	// Handle both GET and POST for /students
	api.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodGet {
			handleStudents(w, r)
		} else if r.Method == http.MethodPost {
			var newStudent Student
			
			// Check if it's JSON request
			if r.Header.Get("Content-Type") == "application/json" {
				if err := json.NewDecoder(r.Body).Decode(&newStudent); err != nil {
					http.Error(w, "Invalid JSON data", http.StatusBadRequest)
					return
				}
			} else {
				// Handle form data
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Invalid form data", http.StatusBadRequest)
					return
				}
				
				newStudent.Name = r.FormValue("name")
				ageStr := r.FormValue("age")
				if ageStr == "" {
					http.Error(w, "Age is required", http.StatusBadRequest)
					return
				}
				age, err := strconv.Atoi(ageStr)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid age: %s (must be a number)", ageStr), http.StatusBadRequest)
					return
				}
				newStudent.Age = age
				newStudent.Email = r.FormValue("email")
			}
			
			// Validate student data
			if err := validateStudent(newStudent); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			mutex.Lock()
			newStudent.ID = len(students) + 1
			students = append(students, newStudent)
			mutex.Unlock()
			
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newStudent)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})


	// GET a specific student by ID
	api.HandleFunc("/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == http.MethodGet {
			id, err := strconv.Atoi(r.PathValue("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			
			mutex.RLock()
			defer mutex.RUnlock()
			
			for _, student := range students {
				if student.ID == id {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(student)
					return
				}
			}
			http.Error(w, "Student not found", http.StatusNotFound)
		} else if r.Method == http.MethodPut {
			// Update a specific student by ID
			id, err := strconv.Atoi(r.PathValue("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			
			var updatedStudent Student
			updatedStudent.ID = id
			
			// Check if it's JSON request
			if r.Header.Get("Content-Type") == "application/json" {
				if err := json.NewDecoder(r.Body).Decode(&updatedStudent); err != nil {
					http.Error(w, "Invalid JSON data", http.StatusBadRequest)
					return
				}
				updatedStudent.ID = id // Ensure ID is set correctly
			} else {
				// Handle form data
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Invalid form data", http.StatusBadRequest)
					return
				}
				
				updatedStudent.Name = r.FormValue("name")
				ageStr := r.FormValue("age")
				if ageStr == "" {
					http.Error(w, "Age is required", http.StatusBadRequest)
					return
				}
				age, err := strconv.Atoi(ageStr)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid age: %s (must be a number)", ageStr), http.StatusBadRequest)
					return
				}
				updatedStudent.Age = age
				updatedStudent.Email = r.FormValue("email")
			}
			
			// Validate student data
			if err := validateStudent(updatedStudent); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			
			// Update the student in the slice
			for i, student := range students {
				if student.ID == updatedStudent.ID {
					students[i] = updatedStudent
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(updatedStudent)
					return
				}
			}
			http.Error(w, "Student not found", http.StatusNotFound)
		} else if r.Method == http.MethodDelete {
			// DELETE a specific student by ID
			id, err := strconv.Atoi(r.PathValue("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			
			mutex.Lock()
			defer mutex.Unlock()
			
			for i, student := range students {
				if student.ID == id {
					students = append(students[:i], students[i+1:]...)
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			http.Error(w, "Student not found", http.StatusNotFound)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Generate summary of a student using Ollama
	api.HandleFunc("/students/{id}/summary", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		
		mutex.RLock()
		var targetStudent *Student
		for _, student := range students {
			if student.ID == id {
				targetStudent = &student
				break
			}
		}
		mutex.RUnlock()
		
		if targetStudent == nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}
		
		// Call Ollama API to generate summary
		summary, err := callOllamaAPI(*targetStudent)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate summary: %v", err), http.StatusInternalServerError)
			return
		}
		
		response := map[string]interface{}{
			"student": targetStudent,
			"summary": summary,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Server starting on port 8000...")
	http.ListenAndServe(":8000", api)
}