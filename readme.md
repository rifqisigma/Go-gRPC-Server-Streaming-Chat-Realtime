# Golang gRPC Server Streaming – Studi Kasus Chat Realtime

Proyek ini adalah implementasi **Realtime Chat** menggunakan **gRPC Server Streaming** dengan bahasa **Go (Golang)**. Fitur utamanya termasuk komunikasi realtime antar pengguna dalam grup, autentikasi dengan JWT, serta penyimpanan data menggunakan MySQL dan GORM.

## 🔧 Teknologi yang Digunakan

- **Go (Golang)**
- **gRPC + Protocol Buffers**
- **MySQL**
- **GORM (ORM)**
- **JWT (JSON Web Token)** untuk autentikasi
- **Server Streaming gRPC** untuk komunikasi realtime
- **Unary RPC** untuk login, register, update, delete, dsb

---

## 🚀 Fitur Utama

- Autentikasi JWT Token via Metadata
- Streaming chat realtime antar member dalam grup
- Unary endpoint untuk pengelolaan user & chat
- Broadcast update & delete melalui stream
- Status online/offline dengan memory connection tracking (dikarenakan bukan didirectional stream, tidak terlalu realtime)


