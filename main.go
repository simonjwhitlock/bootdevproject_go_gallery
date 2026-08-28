package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/database"
	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/httpapi"
	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits      int
	dbQueries           *database.Queries
	tokenSecret         string
	tokenDuration       time.Duration
	refreshTokenTimeout time.Duration
}

func main() {
	fmt.Println("Starting Gallery API...")
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error setting up database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	fmt.Println("Database connected")

	parsedTokenDuration, err := time.ParseDuration(os.Getenv("TOKEN_DEFAULT_DURATION"))
	if err != nil {
		log.Fatal(err)
	}
	parsedRefreshTimeout, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TIMEOUT"))
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		fileserverHits:      0,
		dbQueries:           database.New(db),
		tokenSecret:         os.Getenv("TOKEN_SECRET"),
		tokenDuration:       parsedTokenDuration,
		refreshTokenTimeout: parsedRefreshTimeout,
	}

	// Seed admin user if not exists
	if adminEmail := os.Getenv("ADMIN_EMAIL"); adminEmail != "" {
		if _, err := apiCfg.dbQueries.GetUserByEmail(adminEmail); err == sql.ErrNoRows {
			if adminPassword := os.Getenv("ADMIN_PASSWORD"); adminPassword != "" {
				hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
				now := time.Now()
				_, err := apiCfg.dbQueries.CreateUser(&database.User{
					ID:           uuid.New(),
					CreatedAt:    now,
					UpdatedAt:    now,
					Email:        adminEmail,
					PasswordHash: string(hash),
				})
				if err != nil {
					log.Printf("Warning: failed to create admin user: %v", err)
				} else {
					fmt.Println("Admin user created:", adminEmail)
				}
			}
		}
	}

	// Public handlers
	publicHandler := &httpapi.PublicHandler{DB: apiCfg.dbQueries}

	// Admin handlers
	adminHandler := &httpapi.AdminHandler{
		DB:            apiCfg.dbQueries,
		TokenSecret:   apiCfg.tokenSecret,
		TokenDuration: apiCfg.tokenDuration,
	}

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /api/images", publicHandler.ListImages)
	mux.HandleFunc("GET /api/images/{id}", publicHandler.GetImage)

	// Admin routes (protected)
	mux.HandleFunc("POST /api/admin/login", adminHandler.AdminLogin)
	mux.HandleFunc("GET /api/admin/images", httpapi.AuthMiddleware(apiCfg.tokenSecret, adminHandler.ListImages))
	mux.HandleFunc("POST /api/admin/images", httpapi.AuthMiddleware(apiCfg.tokenSecret, adminHandler.CreateImage))
	mux.HandleFunc("PUT /api/admin/images", httpapi.AuthMiddleware(apiCfg.tokenSecret, adminHandler.UpdateImage))
	mux.HandleFunc("DELETE /api/admin/images", httpapi.AuthMiddleware(apiCfg.tokenSecret, adminHandler.DeleteImage))
	mux.HandleFunc("POST /api/admin/images/reorder", httpapi.AuthMiddleware(apiCfg.tokenSecret, adminHandler.ReorderImages))

	// Serve index.html for non-API routes
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
