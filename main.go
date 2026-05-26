package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"net/mail"
	_ "github.com/go-sql-driver/mysql" // Driver MySQL
)

// Kita buat variabel global 'db' agar bisa diakses oleh semua fungsi
var db *sql.DB
type Mahasiswa struct {
	Nomor int
	NIM int
	Nama string
	Alamat string
	Email string
	NoHP string
	JenisKelamin string
}
var NomorUrut = 1

// Fungsi khusus untuk menyalakan database
func connectDB() {
	var err error
	dsn := "root:@tcp(localhost:3306)/dbl_golang"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("Gagal konfigurasi database: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic("MySQL mati atau tidak ditemukan: " + err.Error())
	}
	fmt.Println("✅ Database MySQL (Laragon) Terhubung!")
}

// Handler 1: Tampilkan Form
func tampilkanForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	filepath := path.Join("views", "form.html")
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler 2: Tangkap, Validasi, dan Simpan ke DB
func prosesPendaftaran(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Harus menggunakan metode POST", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	nama := strings.TrimSpace(r.FormValue("nama"))
	nim := strings.TrimSpace(r.FormValue("nim"))
	alamat := strings.TrimSpace(r.FormValue("alamat"))
	noHp := strings.TrimSpace(r.FormValue("no_hp"))
	email := strings.TrimSpace(r.FormValue("email"))
	jenis_kelamin := r.FormValue("jenis_kelamin")

	// --- SATPAM VALIDASI ---
	if nama == "" || nim == "" || alamat == "" || noHp == "" || email == "" || jenis_kelamin == "" {
		http.Error(w, "Semua field harus diisi!", http.StatusBadRequest)
		return
	}
	if len(nim) < 8 {
		http.Error(w, "Format NIM tidak valid (minimal 8 karakter).", http.StatusBadRequest)
		return
	}
	_, errNim := strconv.Atoi(nim)
	if errNim != nil {
		http.Error(w, "NIM harus berupa angka!", http.StatusBadRequest)
		return
	}
	_, errNoHp := strconv.Atoi(noHp)
	if errNoHp != nil {
		http.Error(w, "No HP harus berupa angka!", http.StatusBadRequest)
		return
	}
	if len(noHp) < 10 {
		http.Error(w, "No HP terlalu pendek (minimal 10 digit).", http.StatusBadRequest)
		return
	}
	_, errEmail := mail.ParseAddress(email)
	if errEmail != nil {
		http.Error(w, "Peringatan: Penulisan alamat email tidak valid!", http.StatusBadRequest)
		return
	}
	
	if jenis_kelamin != "Laki-laki" && jenis_kelamin != "Perempuan" {
		http.Error(w, "Jenis kelamin tidak valid.", http.StatusBadRequest)
		return
	}
	// -----------------------

	// --- PROSES SIMPAN KE DATABASE (GUDANG) ---
	// Kita menggunakan db.Exec untuk menjalankan perintah INSERT
	_, errDB := db.Exec("INSERT INTO tbl_mahasiswa (nama, nim, alamat, email, no_hp, jenis_kelamin) VALUES (?, ?, ?, ?, ?, ?)", nama, nim, alamat, email, noHp, jenis_kelamin)

	if errDB != nil {
		http.Error(w, "Server gagal menyimpan data ke database.", http.StatusInternalServerError)
		fmt.Println("Error DB:", errDB) // Tampilkan error di terminal untuk kita (programmer)
		return
	}

	// Jika sukses melewati DB
	fmt.Fprintf(w, "🎉 PENDAFTARAN BERHASIL!\n\n")
	fmt.Fprintf(w, "Data Saudara %s telah resmi tersimpan di database kami.", nama)
}

// Handler 3: Menampilkan Data Tabel
func tampilkanData(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil data dari database
	rows, err := db.Query("SELECT nim, nama,  alamat, no_hp, email, jenis_kelamin FROM tbl_mahasiswa")
	if err != nil {
		http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Pastikan baris ditutup setelah selesai dibaca

	// 2. Siapkan wadah kosong (Slice/Array) untuk menampung semua data
	var daftarMahasiswa []Mahasiswa

	// 3. Lakukan perulangan untuk membaca data baris demi baris
	for rows.Next() {
		var mhs Mahasiswa
		// Scan bertugas memindahkan data dari database ke dalam Struct mhs
		err = rows.Scan(&mhs.NIM, &mhs.Nama, &mhs.Alamat, &mhs.NoHP, &mhs.Email, &mhs.JenisKelamin)
		if err != nil {
			fmt.Println("Ada error saat scan data:", err)
			continue
		}
		mhs.Nomor = NomorUrut // Beri nomor urut berdasarkan posisi data
		NomorUrut++ // Tambah nomor urut untuk data berikutnya

		// Masukkan data mhs yang sudah terisi ke dalam wadah utama
		daftarMahasiswa = append(daftarMahasiswa, mhs)
	}

	// 4. Kirim wadah utama ke file HTML
	filepath := path.Join("views", "data.html")
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kita mengirim 'daftarMahasiswa' agar bisa dibaca oleh {{range .}} di HTML
	tmpl.Execute(w, daftarMahasiswa) 
}

// Handler 4: Menghapus Data
func hapusData(w http.ResponseWriter, r *http.Request) {
	// 1. Tangkap parameter "nim" dari URL (?nim=...)
	nimTarget := r.URL.Query().Get("nim")

	// Cegah jika ada yang iseng mengakses /hapus tanpa mengirim NIM
	if nimTarget == "" {
		http.Error(w, "NIM tidak ditemukan di URL!", http.StatusBadRequest)
		return
	}

	// 2. Eksekusi perintah DELETE ke MySQL
	// Ingat, kita pakai db.Exec untuk perintah yang mengubah isi database
	_, err := db.Exec("DELETE FROM tbl_mahasiswa WHERE nim = ?", nimTarget)
	
	if err != nil {
		http.Error(w, "Gagal menghapus data dari database", http.StatusInternalServerError)
		fmt.Println("Error Hapus DB:", err)
		return
	}

	// 3. Jika berhasil dihapus, lempar/alihkan kembali pengunjung ke halaman tabel
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}

// Handler 5: Menampilkan Form Edit dengan Data Lama
func tampilkanEditForm(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil NIM dari URL (?nim=...)
	nimTarget := r.URL.Query().Get("nim")
	if nimTarget == "" {
		http.Error(w, "NIM tidak valid", http.StatusBadRequest)
		return
	}

	// 2. Cari data mahasiswa tersebut ke database
	// Kita gunakan db.QueryRow karena kita sudah pasti hanya mengambil SATU data saja
	var mhs Mahasiswa
	err := db.QueryRow("SELECT nim, nama, alamat, no_hp, email, jenis_kelamin FROM tbl_mahasiswa WHERE nim = ?", nimTarget).
		Scan(&mhs.NIM, &mhs.Nama, &mhs.Alamat, &mhs.NoHP, &mhs.Email, &mhs.JenisKelamin)

	if err != nil {
		http.Error(w, "Data mahasiswa tidak ditemukan di database", http.StatusNotFound)
		return
	}

	// 3. Lempar data tersebut ke file edit.html
	filepath := path.Join("views", "edit.html")
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, mhs)
}

// Handler 6: Menyimpan Data Baru Hasil Edit ke Database
func prosesUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Harus menggunakan metode POST", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	// Tangkap data dari form edit
	nim := r.FormValue("nim") // NIM diambil untuk penanda WHERE di SQL
	nama := strings.TrimSpace(r.FormValue("nama"))
	alamat := strings.TrimSpace(r.FormValue("alamat"))
	noHp := strings.TrimSpace(r.FormValue("no_hp"))
	email := strings.TrimSpace(r.FormValue("email"))
	jenisKelamin := r.FormValue("jenis_kelamin")

	// (Kamu bisa menyelipkan satpam validasi noHp dan email di sini seperti biasa)

	// Eksekusi perintah UPDATE ke MySQL
	_, errDB := db.Exec("UPDATE tbl_mahasiswa SET nama=?, alamat=?, no_hp=?, email=?, jenis_kelamin=? WHERE nim=?", 
		nama, alamat, noHp, email, jenisKelamin, nim)

	if errDB != nil {
		http.Error(w, "Gagal memperbarui data di database", http.StatusInternalServerError)
		fmt.Println("Error Update DB:", errDB)
		return
	}

	// Jika sukses, kembalikan ke halaman tabel data
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}





func main() {
	connectDB()
	defer db.Close() 

	http.HandleFunc("/", tampilkanForm)
	http.HandleFunc("/proses", prosesPendaftaran)
	
	// DAFTARKAN RUTE BARU DI SINI
	http.HandleFunc("/data", tampilkanData)
	http.HandleFunc("/hapus", hapusData)
	http.HandleFunc("/edit", tampilkanEditForm)
	http.HandleFunc("/update", prosesUpdate)

	fmt.Println("🚀 Server Web berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
