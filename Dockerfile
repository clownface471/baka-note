# Gunakan image Golang standar (Debian) yang lengkap
FROM golang:1.23

# Set folder kerja
WORKDIR /app

# Copy file resep dependensi dulu (agar dicache oleh Docker)
COPY go.mod go.sum ./

# DOWNLOAD saja, jangan tidy (karena sudah dirapikan lokal)
RUN go mod download

# Copy sisa kode program
COPY . .

# Build aplikasi dengan dukungan CGO (untuk database)
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Buka port
EXPOSE 8080

# Jalankan
CMD ["./main"]