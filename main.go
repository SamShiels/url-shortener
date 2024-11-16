package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type URLShortener struct {
	urls    map[string]string
	baseUrl string
	port    string
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/shorten", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<form method="post" action="/shorten">
				<input type="url" name="url" placeholder="Enter a URL" required>
				<input type="submit" value="Shorten">
			</form>
		</body>
		</html>
	`)
}

func (m *URLShortener) HandleShorten(w http.ResponseWriter, r *http.Request) {
	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortKey := generateShortKey()
	m.urls[shortKey] = originalURL

	shortenedURL := fmt.Sprintf("%s:%s/short/%s", m.baseUrl, m.port, shortKey)

	w.Header().Set("Content-Type", "text/html")
	responseHTML := fmt.Sprintf(`
		<h2>URL Shortener</h2>
		<p>Original URL: %s</p>
		<p>Shortened URL: <a href="%s">%s</a></p>
		<form method="post" action="/shorten">
			<input type="text" name="url" placeholder="Enter a URL">
			<input type="submit" value="Shorten">
		</form>
	`, originalURL, shortenedURL, shortenedURL)

	fmt.Fprint(w, responseHTML)
}

func (m *URLShortener) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[len("/short/"):]
	if shortKey == "" {
		http.Error(w, "Shortened key is missing", http.StatusBadRequest)
		return
	}

	originalURL, found := m.urls[shortKey]
	if !found {
		http.Error(w, "Shortened key not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const keylength = 6

	shortKey := make([]byte, keylength)
	for i := range shortKey {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		shortKey[i] = charset[index.Int64()]
	}

	return string(shortKey)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	shortener := &URLShortener{
		urls:    make(map[string]string),
		baseUrl: baseURL,
		port:    port,
	}

	http.HandleFunc("/", handleForm)
	http.HandleFunc("/shorten", shortener.HandleShorten)
	http.HandleFunc("/short/", shortener.HandleRedirect)
	// Your server setup
	fmt.Println("Base URL:", baseURL)
	fmt.Println("Server running on port:", port)

	fmt.Println("URL Shortener is running on :", port)
	http.ListenAndServe(":"+port, nil)
}
