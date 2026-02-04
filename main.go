package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// --- STRUKTUR DATA ---
type Entry struct {
	ID int `json:"id"`; Content string `json:"content"`; Mood string `json:"mood"`; Author string `json:"author"`; CreatedAt string `json:"created_at"`
}
type BucketItem struct {
	ID int `json:"id"`; Text string `json:"text"`; IsDone bool `json:"is_done"`; Author string `json:"author"`
}
type Memory struct {
	ID int `json:"id"`; Image string `json:"image"`; Caption string `json:"caption"`; Date string `json:"date"`
}
type DeepQ struct {
	Question string `json:"question"`
	AnswerA  string `json:"answer_a"`; AuthorA string `json:"author_a"`
	AnswerB  string `json:"answer_b"`; AuthorB string `json:"author_b"`
}

var db *sql.DB

// --- DATABASE PERTANYAAN AWAL (SEED) ---
// Ini akan dimasukkan ke database saat pertama kali jalan
var starterQuestions = []string{
	"What is a core memory from your childhood that shaped who you are?",
	"Which song instantly reminds you of me, and why?",
	"What is your biggest fear right now?",
	"If we could teleport to anywhere right now, where would we go?",
	"What is one thing you wish people understood better about you?",
	"What does your ideal 'perfect day' look like?",
	"What is a habit of mine that you secretly find cute?",
	"When did you first realize you wanted to be close to me?",
	"What is the most beautiful place you have ever been to?",
	"If you could have dinner with anyone, living or dead, who would it be?",
	"What is a skill you’ve always wanted to learn but haven’t yet?",
	"What is your favorite smell in the world?",
	"If you could change one thing about your past, what would it be?",
	"What makes you feel most loved?",
	"What is the best advice you’ve ever received?",
	"If you had to describe me in three words, what would they be?",
	"What is something small that can always ruin your day?",
	"What is something small that can always make your day better?",
	"Do you believe in soulmates?",
	"What is your biggest regret?",
	"If money wasn't an issue, what job would you do?",
	"What is your favorite way to be comforted when you're sad?",
	"What is the bravest thing you have ever done?",
	"What is a movie that made you cry?",
	"If you could live in any era of history, which one would you pick?",
	"What is your favorite thing about yourself?",
	"What is one goal you want to achieve this year?",
	"If our relationship was a book, what would the title be?",
	"What is the weirdest dream you've ever had?",
	"What are you most grateful for today?",
	"If you could have one superpower, what would it be?",
	"What is a secret talent you have?",
	"What is your favorite memory of us?",
	"If you could see into the future, would you want to?",
	"What is the most valuable lesson life has taught you so far?",
	"What is your definition of success?",
	"If you were an animal, what would you be?",
	"What is your favorite time of day?",
	"What is something you are proud of but rarely talk about?",
	"If you could eat only one meal for the rest of your life, what would it be?",
	"What is your biggest pet peeve?",
	"Who is the most influential person in your life?",
	"What is a book that changed your perspective?",
	"If you were stranded on an island, what 3 things would you bring?",
	"What is the most spontaneous thing you've ever done?",
	"What is your love language?",
	"If you could relive one day of your life, which one would it be?",
	"What is something you want to do before you die?",
	"Do you think you are more like your mom or your dad?",
	"What is your favorite season and why?",
}

func seedQuestions() {
	// Cek apakah tabel kosong
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM question_pool").Scan(&count)
	if err != nil {
		// Jika error (tabel belum ada), abaikan dulu, nanti dibuat di main
		return 
	}
	
	if count == 0 {
		fmt.Println("🌱 Seeding question pool...")
		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("INSERT INTO question_pool (question_text, author) VALUES (?, 'System')")
		for _, q := range starterQuestions {
			stmt.Exec(q)
		}
		stmt.Close()
		tx.Commit()
		fmt.Println("✅ 50+ Questions planted!")
	}
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	dbToken := os.Getenv("DB_TOKEN")
	port := os.Getenv("PORT"); if port == "" { port = "8080" }

	var err error
	db, err = sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", dbUrl, dbToken))
	if err != nil { log.Fatal(err) }
	defer db.Close()

	// 1. Setup Tabel
	db.Exec(`CREATE TABLE IF NOT EXISTS journal (id INTEGER PRIMARY KEY AUTOINCREMENT, content TEXT, mood TEXT, author TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS bucketlist_v2 (id INTEGER PRIMARY KEY AUTOINCREMENT, text TEXT, is_done BOOLEAN DEFAULT 0, author TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS memories (id INTEGER PRIMARY KEY AUTOINCREMENT, image TEXT, caption TEXT, date TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS daily_q_v2 (date TEXT PRIMARY KEY, question TEXT, ans_a TEXT, author_a TEXT, ans_b TEXT, author_b TEXT)`)
	
	// TABEL BARU: KOLAM PERTANYAAN
	db.Exec(`CREATE TABLE IF NOT EXISTS question_pool (id INTEGER PRIMARY KEY AUTOINCREMENT, question_text TEXT, author TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	// Seed Data jika kosong
	seedQuestions()

	// 2. Routing
	mux := http.NewServeMux()
	mux.HandleFunc("/api/journal", handleJournal)
	mux.HandleFunc("/api/bucket", handleBucket)
	mux.HandleFunc("/api/memories", handleMemories)
	mux.HandleFunc("/api/deep", handleDeep)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	fmt.Printf("🚀 Our Space V8.5 (Infinite Questions) :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// --- HANDLERS (Journal, Bucket, Memories SAMA SEPERTI SEBELUMNYA) ---
func handleJournal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		rows, _ := db.Query("SELECT id, content, mood, author, created_at FROM journal ORDER BY created_at DESC LIMIT 50")
		defer rows.Close()
		items := []Entry{}
		for rows.Next() { var i Entry; rows.Scan(&i.ID, &i.Content, &i.Mood, &i.Author, &i.CreatedAt); items = append(items, i) }
		json.NewEncoder(w).Encode(items)
	} else if r.Method == "POST" {
		var i Entry; json.NewDecoder(r.Body).Decode(&i)
		db.Exec("INSERT INTO journal (content, mood, author) VALUES (?, ?, ?)", html.EscapeString(i.Content), i.Mood, html.EscapeString(i.Author))
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func handleBucket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		rows, _ := db.Query("SELECT id, text, is_done, author FROM bucketlist_v2")
		defer rows.Close(); items := []BucketItem{}
		for rows.Next() { var i BucketItem; rows.Scan(&i.ID, &i.Text, &i.IsDone, &i.Author); items = append(items, i) }
		json.NewEncoder(w).Encode(items)
	} else if r.Method == "POST" {
		var i BucketItem; json.NewDecoder(r.Body).Decode(&i)
		if r.URL.Query().Get("action") == "delete" { db.Exec("DELETE FROM bucketlist_v2 WHERE id = ?", i.ID)
		} else if i.ID > 0 { db.Exec("UPDATE bucketlist_v2 SET is_done = ? WHERE id = ?", i.IsDone, i.ID)
		} else { db.Exec("INSERT INTO bucketlist_v2 (text, author) VALUES (?, ?)", html.EscapeString(i.Text), html.EscapeString(i.Author)) }
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func handleMemories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if r.Method == "GET" {
		rows, _ := db.Query("SELECT id, image, caption, date FROM memories ORDER BY id DESC")
		defer rows.Close(); items := []Memory{}
		for rows.Next() { var i Memory; rows.Scan(&i.ID, &i.Image, &i.Caption, &i.Date); items = append(items, i) }
		json.NewEncoder(w).Encode(items)
	} else if r.Method == "POST" {
		var i Memory; json.NewDecoder(r.Body).Decode(&i)
		db.Exec("INSERT INTO memories (image, caption, date) VALUES (?, ?, ?)", i.Image, html.EscapeString(i.Caption), time.Now().Format("2006-01-02"))
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// --- UPDATED DEEP DIVE HANDLER ---
func handleDeep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	today := time.Now().Format("2006-01-02")

	if r.Method == "GET" {
		var q DeepQ
		// 1. Cek apakah hari ini sudah ada pertanyaan?
		err := db.QueryRow(`SELECT question, COALESCE(ans_a,''), COALESCE(author_a,''), COALESCE(ans_b,''), COALESCE(author_b,'') FROM daily_q_v2 WHERE date = ?`, today).Scan(&q.Question, &q.AnswerA, &q.AuthorA, &q.AnswerB, &q.AuthorB)
		
		if err != nil {
			// 2. Jika BELUM ADA, ambil RANDOM dari POOL
			var newQ string
			// ORDER BY RANDOM() LIMIT 1 adalah cara SQL mengambil item acak
			errPool := db.QueryRow("SELECT question_text FROM question_pool ORDER BY RANDOM() LIMIT 1").Scan(&newQ)
			
			if errPool != nil {
				newQ = "What is love?" // Fallback jika pool kosong melompong
			}

			// Simpan pertanyaan terpilih untuk hari ini
			db.Exec("INSERT INTO daily_q_v2 (date, question, ans_a, author_a, ans_b, author_b) VALUES (?, ?, '', '', '', '')", today, newQ)
			q = DeepQ{Question: newQ}
		}
		json.NewEncoder(w).Encode(q)

	} else if r.Method == "POST" {
		// Kita tampung requestnya. Bisa jadi JAWABAN atau PERTANYAAN BARU
		var req struct {
			Type     string `json:"type"`     // 'answer' atau 'new_question'
			Author   string `json:"author"`
			Content  string `json:"content"`  // Bisa berisi jawaban atau teks pertanyaan baru
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Type == "new_question" {
			// LOGIC TAMBAH PERTANYAAN KE POOL
			db.Exec("INSERT INTO question_pool (question_text, author) VALUES (?, ?)", html.EscapeString(req.Content), html.EscapeString(req.Author))
			json.NewEncoder(w).Encode(map[string]string{"status": "added_to_pool"})
			
		} else {
			// LOGIC MENJAWAB PERTANYAAN HARIAN (Seperti sebelumnya)
			var currentA string
			db.QueryRow("SELECT ans_a FROM daily_q_v2 WHERE date = ?", today).Scan(&currentA)
			colAns, colAuth := "ans_a", "author_a"
			if currentA != "" { colAns, colAuth = "ans_b", "author_b" }
			
			query := fmt.Sprintf("UPDATE daily_q_v2 SET %s = ?, %s = ? WHERE date = ?", colAns, colAuth)
			db.Exec(query, html.EscapeString(req.Content), html.EscapeString(req.Author), today)
			json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
		}
	}
}