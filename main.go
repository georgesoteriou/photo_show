package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

type Image struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

type AppState struct {
	sync.RWMutex
	Images     []*Image `json:"images"`
	Interval   int      `json:"interval"`
	Mode       string   `json:"mode"`
	NowShowing string   `json:"now_showing"`
}

var state = AppState{
	Images:     make([]*Image, 0),
	Interval:   5,
	Mode:       "auto-allow", // Default: photos show up immediately
	NowShowing: "",
}

func main() {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// 1. Initial sync to pick up existing files on startup
	syncFolder()

	// 2. Start background worker to watch for manual folder changes
	go func() {
		for {
			time.Sleep(2 * time.Second)
			syncFolder()
		}
	}()

	// --- PORT 8080: PUBLIC UPLOAD FLOW ---
	muxUpload := http.NewServeMux()
	muxUpload.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html, _ := templatesFS.ReadFile("templates/upload.html")
		w.Write(html)
	})
	muxUpload.HandleFunc("/upload", handleUpload)

	go func() {
		fmt.Println("🚀 Public Upload Server running on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", muxUpload))
	}()

	// --- PORT 8081: PRIVATE SHOW & ADMIN FLOW ---
	muxAdmin := http.NewServeMux()
	muxAdmin.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	muxAdmin.HandleFunc("/show", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html, _ := templatesFS.ReadFile("templates/show.html")
		w.Write(html)
	})
	muxAdmin.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html, _ := templatesFS.ReadFile("templates/admin.html")
		w.Write(html)
	})

	// API Endpoints
	muxAdmin.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		state.RLock()
		defer state.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	})

	muxAdmin.HandleFunc("/api/toggle-image", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Filename string `json:"filename"` }
		json.NewDecoder(r.Body).Decode(&req)

		state.Lock()
		for _, img := range state.Images {
			if img.Filename == req.Filename {
				if img.Status == "allowed" {
					img.Status = "denied"
				} else {
					img.Status = "allowed"
				}
				break
			}
		}
		state.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	muxAdmin.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Interval int    `json:"interval"`
			Mode     string `json:"mode"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		state.Lock()
		if req.Interval > 0 {
			state.Interval = req.Interval
		}
		// Updated validation to use the new mode names
		if req.Mode == "auto-allow" || req.Mode == "auto-deny" {
			state.Mode = req.Mode
		}
		state.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	muxAdmin.HandleFunc("/api/now-showing", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Filename string `json:"filename"` }
		json.NewDecoder(r.Body).Decode(&req)
		
		state.Lock()
		state.NowShowing = req.Filename
		state.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("🔒 Private Admin/Show Server running on http://localhost:8081")
	fmt.Println("   - Display: http://localhost:8081/show")
	fmt.Println("   - Admin:   http://localhost:8081/admin")
	log.Fatal(http.ListenAndServe(":8081", muxAdmin))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	outPath := filepath.Join("uploads", filename)

	out, err := os.Create(outPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	syncFolder() 

	http.Redirect(w, r, "/?success=1", http.StatusSeeOther)
}

func syncFolder() {
	entries, err := os.ReadDir("uploads")
	if err != nil {
		return
	}

	var files []os.FileInfo
	diskMap := make(map[string]bool)
	
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil && info.Size() > 0 {
				files = append(files, info)
				diskMap[entry.Name()] = true
			}
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	state.Lock()
	defer state.Unlock()

	var newImages []*Image
	known := make(map[string]bool)
	
	for _, img := range state.Images {
		if diskMap[img.Filename] {
			newImages = append(newImages, img)
			known[img.Filename] = true
		} else if img.Filename == state.NowShowing {
			state.NowShowing = ""
		}
	}
	state.Images = newImages

	for _, f := range files {
		if !known[f.Name()] {
			status := "allowed"
			// Now checks for the new "auto-deny" state
			if state.Mode == "auto-deny" {
				status = "denied"
			}
			state.Images = append(state.Images, &Image{
				Filename: f.Name(),
				Status:   status,
			})
		}
	}
}