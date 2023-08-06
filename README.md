---

# Online Store API

Online Store API adalah sebuah aplikasi RESTful yang memungkinkan pengguna untuk menjelajahi daftar produk, mengelola keranjang belanja, melakukan checkout, dan mengelola akun pengguna.

## Daftar Isi

- [Fitur](#fitur)
- [Persyaratan](#persyaratan)
- [Instalasi](#instalasi)
- [Penggunaan](#penggunaan)
- [Endpoints](#endpoints)
- [Lisensi](#lisensi)

## Fitur

Aplikasi Online Store API memiliki beberapa fitur utama:

- Menampilkan daftar produk dengan informasi rinci.
- Pengguna dapat melihat keranjang belanja mereka.
- Pengguna dapat menambahkan produk ke dalam keranjang belanja.
- Pengguna dapat menghapus produk dari keranjang belanja.
- Pengguna dapat melakukan checkout dan membuat pesanan.
- Pengguna dapat mendaftar dan masuk menggunakan akun pengguna.
- Akses tervalidasi menggunakan JWT (JSON Web Token).

## Persyaratan

Sebelum Anda memulai, pastikan Anda telah memenuhi persyaratan berikut:

- Go (minimal versi 1.16) terinstal di sistem Anda.
- MySQL database telah terpasang dan berjalan.
- Paket Fiber dan GORM terinstal. Anda dapat menginstalnya dengan perintah `go get github.com/gofiber/fiber/v2 github.com/jinzhu/gorm`.

## Instalasi

Berikut adalah langkah-langkah untuk menginstal dan menjalankan aplikasi:

1. Salin repositori ini dengan perintah:

   ```sh
   git clone https://github.com/zaentaqin/online-store.git
   cd online-store-api
   ```

2. Salin file `.env.example` menjadi `.env` dan sesuaikan pengaturan database Anda.

3. Jalankan perintah berikut untuk menginstal dependensi:

   ```sh
   go mod download
   ```

4. Jalankan aplikasi dengan perintah:

   ```sh
   go run main.go
   ```

Aplikasi akan berjalan di `http://localhost:8080`.

## Penggunaan

Untuk menggunakan API ini, Anda dapat menggunakan alat pengujian API seperti [Postman](https://www.postman.com/) atau `curl` di terminal. Pastikan untuk melampirkan token JWT yang valid pada setiap permintaan yang memerlukannya.

## Endpoints

Berikut adalah daftar endpoint yang tersedia:

- `GET /api/products`: Mendapatkan daftar produk.
- `GET /api/products/:id`: Mendapatkan detail produk berdasarkan ID.
- `GET /api/shopping-cart`: Melihat keranjang belanja pengguna.
- `POST /api/shopping-cart`: Menambah produk ke dalam keranjang belanja.
- `DELETE /api/shopping-cart/:id`: Menghapus produk dari keranjang belanja.
- `POST /api/checkout`: Melakukan proses checkout dan membuat pesanan.
- `POST /api/register`: Mendaftar akun pengguna baru.
- `POST /api/login`: Masuk ke akun pengguna dan mendapatkan token JWT.

## Lisensi

Proyek ini dilisensikan di bawah Lisensi MIT - lihat berkas [LISENSI](LISENSI) untuk detail lebih lanjut.

---
