# Transjakarta Fleet Backend (Golang)

## Alur Kerja (Workflow)

```
Publisher (Go) --> MQTT Broker (Mosquitto) --> Backend (Go)
                                               |
                                               |--> PostgreSQL (Store locations)
                                               |
                                               |--> RabbitMQ (Geofence events) --> Worker (Go) (Logs alerts)
                                               |
                                               |--> API (Gin) (Serves location data)
```

1. **Publisher**: Mengirim data lokasi kendaraan ke MQTT broker setiap 2 detik.
2. **MQTT Broker**: Menerima dan mendistribusikan pesan lokasi.
3. **Backend**:
   - Subscribe ke MQTT, validasi dan simpan data ke PostgreSQL.
   - Cek apakah kendaraan dalam geofence (radius 50m dari titik tertentu).
   - Jika ya, kirim event geofence ke RabbitMQ.
   - Serve API untuk query lokasi kendaraan.
4. **Worker**: Konsumsi event geofence dari RabbitMQ dan log ke console.
5. **PostgreSQL**: Penyimpanan data lokasi kendaraan.
6. **RabbitMQ**: Queue untuk event geofence.

##  Sistem Manajemen Armada

### Deskripsi Proyek
Sistem backend untuk manajemen armada Transjakarta yang dapat:
1. Menerima data lokasi kendaraan melalui MQTT
2. Menyimpan data lokasi ke PostgreSQL
3. Menyediakan API untuk mendapatkan lokasi terakhir dan riwayat perjalanan kendaraan
4. Menggunakan RabbitMQ untuk event geofence
5. Menggunakan Docker untuk deployment

### Teknologi yang Digunakan
- **Golang**: Pengembangan layanan backend
- **MQTT (Eclipse Mosquitto)**: Menerima data lokasi kendaraan
- **PostgreSQL**: Penyimpanan data lokasi kendaraan
- **RabbitMQ**: Event processing berbasis geofence
- **Docker**: Containerisasi seluruh sistem

### Spesifikasi Teknis

#### 1. Menerima Data Lokasi Kendaraan via MQTT
- Subscriber MQTT mendengarkan topik: `/fleet/vehicle/{vehicle_id}/location`
- Format data JSON:
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```
- Validasi format data dilakukan

#### 2. Penyimpanan Data Lokasi ke PostgreSQL
- Tabel: `vehicle_locations`
- Field: `vehicle_id`, `latitude`, `longitude`, `timestamp`
- Service Golang untuk insert data ke PostgreSQL

#### 3. API untuk Mengakses Data Lokasi
- Framework: Gin
- Endpoints:
  - `GET /vehicles/{vehicle_id}/location` - Lokasi terakhir
  - `GET /vehicles/{vehicle_id}/history?start=...&end=...` - Riwayat dalam rentang waktu

#### 4. RabbitMQ untuk Event Geofence
- Trigger event saat kendaraan masuk radius 50 meter dari titik tertentu
- Konfigurasi:
  - Exchange: `fleet.events`
  - Queue: `geofence_alerts`
- Format pesan:
```json
{
  "vehicle_id": "B1234XYZ",
  "event": "geofence_entry",
  "location": {
    "latitude": -6.2088,
    "longitude": 106.8456
  },
  "timestamp": 1715003456
}
```
- Worker service membaca dari queue `geofence_alerts`

#### 5. Docker & Deployment
- Docker Compose mencakup:
  - Backend Golang
  - PostgreSQL
  - RabbitMQ
  - MQTT Broker (Eclipse Mosquitto)
- Dockerfile untuk aplikasi Golang

### Persyaratan Tambahan
- Script publisher mock dalam Go
- Mengirim data ke MQTT topik `/fleet/vehicle/{vehicle_id}/location` setiap 2 detik

## Cara Menjalankan

### Lokal dengan Docker Compose
1. Copy `.env.example` ke `.env` dan sesuaikan jika perlu
2. Jalankan `docker compose up --build`
3. Backend tersedia di `http://localhost:8080`

### Menjalankan Mock Publisher
```
go run cmd/publisher/publisher.go B1234XYZ
```

## Testing

### Cara Menjalankan Test
1. Pastikan Docker Compose berjalan untuk database dan message broker
2. Jalankan `go test ./...` untuk unit test
3. Untuk integration test, pastikan PostgreSQL dan RabbitMQ tersedia

### Hasil yang Diharapkan
- **Publikasi MQTT**: Data lokasi berhasil dikirim
- **Penyimpanan PostgreSQL**: Data tersimpan dengan benar
- **API Lokasi Terakhir**: Endpoint mengembalikan data benar
- **API Riwayat**: Mengembalikan data dalam rentang waktu
- **Geofence Event**: Pesan dikirim ke RabbitMQ saat masuk geofence
- **Docker Compose**: Semua layanan berjalan lancar

## Catatan
- Jalankan `sql/create_tables.sql` pada database PostgreSQL sebelum testing
