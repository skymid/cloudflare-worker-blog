package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//var videoDir = "./"

var videoDir = "F:\\xz"

func getVideoListHandler(w http.ResponseWriter, r *http.Request) {
	var videoFiles []string

	err := filepath.Walk(videoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			relPath, err := filepath.Rel(videoDir, path)
			if err != nil {
				return err
			}
			videoFiles = append(videoFiles, relPath)
		}
		return nil
	})

	if err != nil {
		log.Println("Error walking the video directory:", err)
		http.Error(w, "Failed to retrieve video list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videoFiles)
}
func saveFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		List []string `json:"list"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 检查 list 是否为空
	if len(data.List) == 0 {
		log.Println("Received empty list")
		http.Error(w, "List is empty", http.StatusBadRequest)
		return
	}

	// 将 list 转换为文本内容，每行一个元素
	content := strings.Join(data.List, "\n") + "\n"

	// 保存到文件（例如 favorites.txt）
	filePath := filepath.Join(videoDir, "favorites.txt")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Println("Error writing to file:", err)
		http.Error(w, "Failed to save list to file", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "List saved successfully",
		"file":    filePath,
	})
}
func loadFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(videoDir, "favorites.txt")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，返回空数组
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading favorites file:", err)
		http.Error(w, "Failed to read favorites file", http.StatusInternalServerError)
		return
	}

	// 将文件内容按行分割为数组，去除空行
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	var favorites []string
	for _, line := range lines {
		if line != "" { // 忽略空行
			favorites = append(favorites, line)
		}
	}

	// 返回 JSON 格式的数组
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}
func main() {

	if _, err := os.Stat(videoDir); os.IsNotExist(err) {
		err := os.MkdirAll(videoDir, 0755)
		if err != nil {
			log.Fatalf("Error creating video directory: %v", err)
		}
		fmt.Printf("Created video directory: %s\n", videoDir)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/videos", getVideoListHandler).Methods("GET")
	router.HandleFunc("/api/saveFavorites", saveFavoritesHandler).Methods("POST")
	router.HandleFunc("/api/loadFavorites", loadFavoritesHandler).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(videoDir)))

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	port := ":8080"
	log.Printf("Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
