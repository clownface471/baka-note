# 1. Gunakan 'Base Image' Go yang ringan
FROM golang:1.23-alpine

# 2. Set folder kerja di dalam "komputer" Koyeb
WORKDIR /app

# 3. Copy daftar belanjaan (library) dulu biar cepat
COPY go.mod go.sum ./
RUN go mod download

# 4. Copy seluruh kode program
COPY . .

# 5. Rakit (Build) aplikasinya jadi file bernama 'main'
RUN go build -o main .

# 6. Buka pintu port 8080
EXPOSE 8080

# 7. Jalankan aplikasinya
CMD ["./main"]