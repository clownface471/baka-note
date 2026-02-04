# Gunakan image Golang standar (Debian) yang punya tools lengkap
FROM golang:1.23

# Set folder kerja
WORKDIR /app

# Copy seluruh kode (termasuk go.mod dan go.sum)
COPY . .

# Download dependency & Pastikan rapi
RUN go mod tidy
RUN go mod download

# Build aplikasinya
# CGO_ENABLED=1 penting untuk beberapa library database
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Buka port
EXPOSE 8080

# Jalankan
CMD ["./main"]