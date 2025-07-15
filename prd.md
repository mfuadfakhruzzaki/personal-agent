Product Requirements Document (PRD)
AI-Powered Todo Agent Backend (Go)

⸻

1. Latar Belakang & Visi

Visi:
Membangun backend ringan, hemat resource, dan scalable untuk sebuah AI-powered “Todo Agent” yang dapat menerima input teks, gambar, atau dokumen dari pengguna, memprosesnya via Gemini API, lalu men-push daftar tugas (todo list) ke Supabase, dan di-fetch oleh frontend Next.js.

Masalah yang Diselesaikan:
• Memudahkan pengguna mengubah catatan bebas (text, foto, PDF) menjadi daftar tugas terstruktur.
• Menyediakan pipeline automasi end-to-end tanpa harus self-host model LLM.
• Menjamin penggunaan AWS EC2 Free Tier (1 vCPU, 1 GB RAM) tetap lancar.

⸻

2. Sasaran & Key Metrics

Sasaran Utama Key Metrics

1. Keandalan: API ≥ 99% uptime • Uptime ≥ 99% (monitored via Health Check)
2. Performansi: Respon ≤ 2 detik • P95 latency < 2 detik
3. Resource Efficiency • RAM idle ≤ 50 MB; peak ≤ 400 MB
4. Scalability Modular • Mudah tambah service (OCR, parser) terpisah
5. Developer Experience • Dokumen & contoh kode lengkap, CI/CD jalan

⸻

3. Fitur Utama
   1. Endpoint /process
      • Method: POST
      • Payload multipart/form-data:
      • type: "text" | "image" | "document"
      • content: teks langsung (jika type=text)
      • file: upload file (jika type=image|document)
      • user_id: string
   2. Pre-processing
      • Jika image: unggah sementara, kirim ke Gemini Vision API.
      • Jika document: parsing via pdfcpu/python-docx + (opsional) Tesseract OCR lokal.
   3. AI Parsing
      • Forward teks ke Gemini API dengan prompt JSON-todo (see Prompt Template).
      • Terima JSON array [{"title","description","due_date"}].
   4. Persistence
      • Insert record ke Supabase table todos dengan kolom:
      • id (UUID), user_id, title, description, due_date, source_type, source_url, created_at.
   5. Response
      • Mengembalikan 202 Accepted dengan job_id.
      • Endpoint polling /status/{job_id} untuk cek status & hasil jika diperlukan.
   6. Background Worker
      • Proses parsing + persistence asinkron dengan goroutine + job queue sederhana (Redis minimal atau channel Go).

⸻

4. User Stories 1. User As a Textual Input
   “Sebagai pengguna, saya ingin mengirim teks bebas agar sistem mengubahnya menjadi daftar tugas terstruktur.” 2. User As an Image Input
   “Sebagai pengguna, saya ingin mengirim foto catatan tangan agar di-OCR & di-parse menjadi todo list.” 3. User As a Document Input
   “Sebagai pengguna, saya ingin mengupload file PDF berisi catatan rapat agar tercipta daftar tugas.” 4. Developer As Observer
   “Sebagai developer, saya perlu logging terstruktur agar mudah debugging dan monitoring resource.”

⸻

5. Functional Requirements
   • FR1: Sistem harus menerima request text/image/document pada /process.
   • FR2: Sistem memvalidasi ukuran file ≤ 5 MB; menolak jika lebih besar.
   • FR3: Sistem harus terhubung ke Gemini API (Key via env var).
   • FR4: Sistem mencatat setiap request & response time ke logging.
   • FR5: Push data todo ke Supabase via Go SDK.
   • FR6: Sediakan endpoint /healthz untuk health check.
   • FR7: Configurable via config.yaml (API keys, timeouts, swap path, dsb).

⸻

6. Non-Functional Requirements
   • NFR1: Memory footprint idle ≤ 50 MB, peak ≤ 400 MB.
   • NFR2: P95 latency total end-to-end < 2 detik (tanpa hit Gemini).
   • NFR3: Uptime ≥ 99% dalam periode 30 hari.
   • NFR4: Otentikasi request via HMAC signature atau API key header.
   • NFR5: Containerized (Docker), mudah deploy di EC2 Free Tier.
   • NFR6: Graceful shutdown: drain jobs sebelum SIGTERM.
   • NFR7: Retry otomatis pada gagal network (max 3 retry expo backoff).

⸻

7. Teknis & Arsitektur

7.1 Tech Stack
• Bahasa & Framework: Go (1.21+) + Gin (atau Fiber)
• Queue: Go channel (memadai) atau Redis (opsional)
• Database: Supabase (PostgreSQL + Storage)
• Logging: zap atau logrus (structured JSON)
• OCR (opsional): Tesseract (caveat RAM)
• Container: Docker + Docker Compose
• Deployment: AWS EC2 Free Tier (t2.micro/t3.micro)
• CI/CD: GitHub Actions → Docker build & push → deploy via SSH & systemd

7.2 Diagram Arsitektur

[User Frontend Next.js]
└─(HTTPS POST /process)→ [Go Backend EC2]
├─ Pre-process (OCR/PDF)
├─ Gemini API Call
├─ Push → Supabase
└─ Respond job_id

[Supabase] ←─(Data Layer)→

⸻

8. API Design

Endpoint Method Auth Request Response
/process POST API-Key multipart/form-data: type, content/file, user_id 202 Accepted + {job_id}
/status/{job_id} GET API-Key path param job_id `{ status: “pending
/healthz GET — — 200 OK

Prompt Template (Gemini):

Anda adalah asisten produktivitas. Dari teks berikut, ekstrak daftar todo dalam format JSON:
[{"title":"…","description":"…","due_date":"YYYY-MM-DD|null"}]
Teks:

---

## {{parsed_text}}

⸻

9. Data Model (Supabase)

CREATE TABLE todos (
id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
user_id text NOT NULL,
title text NOT NULL,
description text,
due_date date,
source_type text NOT NULL,
source_url text,
created_at timestamptz DEFAULT now()
);

⸻

10. Milestones & Timeline

Milestone Durasi Tanggal Target
Setup project repo & CI 1 hari 17 Jul 2025
Implement /process basic 2 hari 19 Jul 2025
Integrasi Gemini API + prompt 2 hari 21 Jul 2025
Supabase integration & tests 2 hari 23 Jul 2025
Error handling & logging 1 hari 24 Jul 2025
Dockerize + systemd deploy script 1 hari 25 Jul 2025
Load test & resource tuning 2 hari 27 Jul 2025
Dokumentasi & handover 1 hari 28 Jul 2025

⸻

11. Keamanan & Operasional
    • API Key: Simpan di AWS Secrets Manager / EC2 env.
    • Rate limiting: max 5 req/detik per IP.
    • Swap: atur swapfile 512 MB untuk mencegah OOM.
    • Monitoring: gunakan CloudWatch untuk CPU/RAM, alert > 80%.
    • Backup: Supabase otomatis daily; simpan logs ke S3.

⸻

12. Kesimpulan

Dokumen ini merangkum seluruh kebutuhan fungsional, non-fungsional, arsitektur, dan timeline. Backend Go ini akan ringan, hemat resource, dan terintegrasi mulus dengan Gemini API & Supabase, cocok dijalankan pada AWS EC2 Free Tier.

⸻
