package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"           // Import sync to use synchronization primitives like RWMutex.
)

// Book struct defines the model for storing book data.
type Book struct {
    ID    string `json:"id"`    // ID as string, used as a unique identifier for books.
    Title string `json:"title"` // Title of the book.
}

var (
    books = make(map[string]Book) // Map to store books with their ID as the key.
    mux   sync.RWMutex            // RWMutex to safeguard the books map for concurrent access.
)

func main() {
	// Initialize default books
    initializeBooks()

    // Create a new HTTP server
    server := &http.Server{
        Addr:    ":8080",
        Handler: nil, // Use the default ServeMux
    }

    // Set up HTTP routes
    http.HandleFunc("/books", authenticate(handleBooks))
    http.HandleFunc("/book/", authenticate(handleBook))

    // Start the HTTP server in a separate goroutine so that it doesn't block.
    go func() {
        fmt.Println("Server starting on port 8080...")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe(): %v", err)
        }
    }()

    // Listen for interrupt signal to gracefully shut down the server
    quit := make(chan os.Signal, 1)
    // Trigger graceful shutdown on interrupt signals
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Context to tell the server it has 5 seconds to finish
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Shutting down the server
    fmt.Println("Shutting down server...")
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
}

func initializeBooks() {
    books["1"] = Book{ID: "1", Title: "1984"}
    books["2"] = Book{ID: "2", Title: "Brave New World"}
    books["3"] = Book{ID: "3", Title: "To Kill a Mockingbird"}
    books["4"] = Book{ID: "4", Title: "The Great Gatsby"}
    books["5"] = Book{ID: "5", Title: "Moby Dick"}
}

// authenticate is a middleware function that verifies the presence of an API key.
func authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key") // Retrieve the API key from the header.
        if apiKey != "secret-key" {         // Check if the provided API key matches the expected value.
            http.Error(w, "Unauthorized", http.StatusUnauthorized) // Send an unauthorized status if the key does not match.
            return
        }
        next(w, r) // Call the next handler if the API key is valid.
    }
}

// handleBooks handles requests for the /books route.
func handleBooks(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET": // Handle GET requests to retrieve all books.
        mux.RLock() // Read-lock the mutex before accessing the shared map.
        bks := make([]Book, 0, len(books)) // Create a slice of books to send back.
        for _, book := range books {
            bks = append(bks, book) // Append each book to the slice.
        }
        mux.RUnlock() // Unlock the mutex after reading.
        json.NewEncoder(w).Encode(bks) // Send the books as JSON.

    case "POST": // Handle POST requests to add new books.
        var book Book
        if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest) // Send an error if the book cannot be decoded.
            return
        }
        mux.Lock()              // Lock the mutex before modifying the map.
        books[book.ID] = book  // Add the book to the map.
        mux.Unlock()            // Unlock the mutex after modifying.
        w.WriteHeader(http.StatusCreated) // Respond with a status indicating creation.

    default:
        w.WriteHeader(http.StatusMethodNotAllowed) // Send an error if the method is not supported.
    }
}

// handleBook handles requests for the /book/{id} route.
func handleBook(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/book/"):] // Extract the book ID from the URL path.
    switch r.Method {
    case "GET": // Handle GET requests to retrieve a single book by ID.
        mux.RLock()            // Read-lock the mutex before accessing the map.
        book, ok := books[id]  // Retrieve the book from the map.
        mux.RUnlock()          // Unlock the mutex after accessing.
        if !ok {
            http.NotFound(w, r) // If the book is not found, send a 404 response.
            return
        }
        json.NewEncoder(w).Encode(book) // Send the book as JSON.

    case "PUT": // Handle PUT requests to update an existing book.
        var book Book
        if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest) // Send an error if the book cannot be decoded.
            return
        }
        mux.Lock()             // Lock the mutex before modifying the map.
        books[id] = book      // Update the book in the map.
        mux.Unlock()           // Unlock the mutex after modifying.
        json.NewEncoder(w).Encode(book) // Send the updated book as JSON.

    case "DELETE": // Handle DELETE requests to remove a book by ID.
        mux.Lock()            // Lock the mutex before modifying the map.
        delete(books, id)     // Remove the book from the map.
        mux.Unlock()          // Unlock the mutex after modifying.
        w.WriteHeader(http.StatusNoContent) // Send a status to indicate successful deletion.

    default:
        w.WriteHeader(http.StatusMethodNotAllowed) // Send an error if the method is not supported.
    }
}