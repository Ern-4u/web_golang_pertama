package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
)

// Handler 1: Menampilkan halaman form
func tampilkanForm(w http.ResponseWriter, r *http.Request) {
	// Pastikan hanya melayani request GET biasa (ketika orang baru membuka URL)
	if r.Method != "GET" {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	filepath := path.Join("views", "index.html")
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// Handler 2: Memproses data yang dikirim dari form
func prosesPendaftaran(w http.ResponseWriter, r *http.Request) {
	// Wajibkan harus menggunakan metode POST!
	if r.Method != "POST" {
		http.Error(w, "Harus menggunakan metode POST", http.StatusMethodNotAllowed)
		return
	}

	// 1. Wajib dipanggil sebelum mengambil data! Ini menyuruh Go mengurai isi form.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Gagal memproses form", http.StatusInternalServerError)
		return
	}

	// 2. Ambil datanya berdasarkan atribut 'name' di file HTML
	namaPengirim := r.FormValue("nama")
	nimPengirim := r.FormValue("nim")
	pilihanDivisi := r.FormValue("divisi")

	// 3. Cetak hasilnya sebagai balasan ke layar browser
	fmt.Fprintf(w, "Pendaftaran Sukses!\n\n")
	fmt.Fprintf(w, "Selamat bergabung, %s (%s).\n", namaPengirim, nimPengirim)
	fmt.Fprintf(w, "Anda telah terdaftar di Divisi %s.", pilihanDivisi)
}

func main() {
	// Rute untuk melihat form
	http.HandleFunc("/", tampilkanForm)
	
	// Rute tujuan dari 'action' di form HTML
	http.HandleFunc("/proses", prosesPendaftaran)

	fmt.Println("Server berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}