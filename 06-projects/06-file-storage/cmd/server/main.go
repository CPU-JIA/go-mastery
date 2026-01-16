package main

import (
	"log"
	"net/http"
	"os"

	"file-storage-service/internal/config"
	"file-storage-service/internal/handlers"
	"file-storage-service/internal/services"
	"file-storage-service/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化存储
	fileStorage, err := storage.NewFileStorage(cfg.Storage)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}

	// 初始化服务
	fileService, err := services.NewFileService(fileStorage, cfg)
	if err != nil {
		log.Fatal("Failed to initialize file service:", err)
	}

	// 初始化处理器
	fileHandler := handlers.NewFileHandler(fileService)

	// 设置路由
	router := mux.NewRouter()

	// API 路由
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/files", fileHandler.UploadFile).Methods("POST")
	api.HandleFunc("/files", fileHandler.ListFiles).Methods("GET")
	api.HandleFunc("/files/{id}", fileHandler.GetFile).Methods("GET")
	api.HandleFunc("/files/{id}", fileHandler.DeleteFile).Methods("DELETE")
	api.HandleFunc("/files/{id}/download", fileHandler.DownloadFile).Methods("GET")
	api.HandleFunc("/upload-token", fileHandler.GenerateUploadToken).Methods("POST")

	// 静态文件路由
	router.HandleFunc("/", fileHandler.ServeUploadPage).Methods("GET")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("./web/static/"))))

	// 中间件
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	log.Printf("File Storage Service starting on port %s", port)
	log.Printf("Storage provider: %s", cfg.Storage.Provider)
	log.Printf("Max file size: %d MB", cfg.Upload.MaxSize/(1024*1024))

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		}

		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}

		if originAllowed || origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Upload-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
