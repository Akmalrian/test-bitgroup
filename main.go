package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Produk struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Harga int    `json:"harga"`
}

var (
	db     = make(map[int]Produk)
	nextID = 1
	mu     sync.Mutex
)

func main() {
	db[1] = Produk{ID: 1, Nama: "Laptop", Harga: 8500000}
	db[2] = Produk{ID: 2, Nama: "Mouse", Harga: 150000}
	nextID = 3

	http.HandleFunc("/api/produk", handleProduk)
	http.HandleFunc("/api/produk/", handleProdukByID)

	fmt.Println("server berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleProduk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		daftar := make([]Produk, 0, len(db))
		for _, v := range db {
			daftar = append(daftar, v)
		}
		mu.Unlock()
		json.NewEncoder(w).Encode(daftar)

	case http.MethodPost:
		var p Produk
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
			return
		}

		mu.Lock()
		p.ID = nextID
		nextID++
		db[p.ID] = p
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}

func handleProdukByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID harus berupa angka", http.StatusBadRequest)
		return
	}

	mu.Lock()
	p, exists := db[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Produk tidak ditemukan", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(p)

	case http.MethodPut:
		var pBaru Produk
		if err := json.NewDecoder(r.Body).Decode(&pBaru); err != nil {
			http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
			return
		}

		mu.Lock()
		pBaru.ID = id
		db[id] = pBaru
		mu.Unlock()

		json.NewEncoder(w).Encode(pBaru)

	case http.MethodDelete:
		mu.Lock()
		delete(db, id)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Berhasil menghapus"})

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}
