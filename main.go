package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings" // Tambahkan ini
	"strconv" // Tambahkan ini
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



func prosesPendaftaran(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Harus menggunakan metode POST", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	// 1. Ambil data dan bersihkan spasi yang tidak sengaja terketik di awal/akhir
	namaPengirim := strings.TrimSpace(r.FormValue("nama"))
	nimPengirim := strings.TrimSpace(r.FormValue("nim"))
	pilihanDivisi := r.FormValue("divisi")

	// ==========================================
	// 2. MULAI PROSES VALIDASI
	// ==========================================

	// Validasi A: Pastikan tidak ada yang kosong
	if namaPengirim == "" || nimPengirim == "" {
		// Kita kembalikan pesan error dan hentikan proses dengan 'return'
		http.Error(w, "Peringatan: Nama dan NIM tidak boleh kosong!", http.StatusBadRequest)
		return
	}

	// Validasi B: Pastikan NIM panjangnya minimal 8 karakter
	if len(nimPengirim) < 8 {
		http.Error(w, "Peringatan: Format NIM tidak valid (minimal 8 karakter).", http.StatusBadRequest)
		return
	}

	// Validasi C: Pastikan NIM HANYA berisi angka
	// strconv.Atoi mencoba mengubah string menjadi integer (angka bulat).
	// Jika gagal (error tidak nil), berarti ada huruf di dalamnya.
	_, errAngka := strconv.Atoi(nimPengirim)
	if errAngka != nil {
		http.Error(w, "Peringatan: NIM harus berupa angka!", http.StatusBadRequest)
		return
	}

	// Validasi D: Pastikan Divisi sesuai dengan opsi panitia yang ada
	if pilihanDivisi != "Acara" && pilihanDivisi != "Humas" && pilihanDivisi != "Perlengkapan" {
		http.Error(w, "Peringatan: Pilihan divisi tidak dikenali.", http.StatusBadRequest)
		return
	}

	// ==========================================
	// 3. JIKA LOLOS SEMUA VALIDASI, TAMPILKAN SUKSES
	// ==========================================
	
	fmt.Fprintf(w, "PENDAFTARAN BERHASIL!\n\n")
	fmt.Fprintf(w, "Nama   : %s\n", namaPengirim)
	fmt.Fprintf(w, "NIM    : %s\n", nimPengirim)
	fmt.Fprintf(w, "Divisi : %s\n", pilihanDivisi)
	fmt.Fprintf(w, "\nData Anda siap dimasukkan ke dalam database panitia.")
}


func main() {
	// Rute untuk melihat form
	http.HandleFunc("/", tampilkanForm)
	
	// Rute tujuan dari 'action' di form HTML
	http.HandleFunc("/proses", prosesPendaftaran)

	fmt.Println("Server berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}