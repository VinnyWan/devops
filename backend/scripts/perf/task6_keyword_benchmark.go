package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type benchmarkResult struct {
	label      string
	plan       []string
	durations  []time.Duration
	average    time.Duration
	maximum    time.Duration
	minimum    time.Duration
	resultRows int
}

func main() {
	var (
		rowCount = flag.Int("rows", 5_000_000, "压测数据量")
		runs     = flag.Int("runs", 5, "每个查询重复次数")
	)
	flag.Parse()

	dbPath := filepath.Join("tmp", "task6_keyword_benchmark.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		panic(err)
	}
	_ = os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	execMust(db, "PRAGMA journal_mode = WAL;")
	execMust(db, "PRAGMA synchronous = OFF;")
	execMust(db, "PRAGMA temp_store = MEMORY;")
	execMust(db, "PRAGMA cache_size = -200000;")

	setupSchema(db)
	seedData(db, *rowCount)

	baselineSQL := `
SELECT id, username, name, email
FROM users
WHERE env = ?
  AND (LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ?)
ORDER BY created_at DESC
LIMIT 20;
`
	optimizedSQL := `
SELECT id, username, name, email
FROM users
WHERE env = ?
  AND keyword_index LIKE ?
ORDER BY created_at DESC
LIMIT 20;
`

	baselinePlan := explainQueryPlan(db, baselineSQL, "prod", "%sre-ops-target%", "%sre-ops-target%", "%sre-ops-target%")
	baseline := runBenchmark(db, "baseline_like", baselineSQL, *runs, "prod", "%sre-ops-target%", "%sre-ops-target%", "%sre-ops-target%")
	baseline.plan = baselinePlan

	execMust(db, "CREATE INDEX idx_users_env_keyword_created ON users (env, keyword_index, created_at DESC, id);")

	optimizedPlan := explainQueryPlan(db, optimizedSQL, "prod", "sre-ops-target%")
	optimized := runBenchmark(db, "optimized_prefix_index", optimizedSQL, *runs, "prod", "sre-ops-target%")
	optimized.plan = optimizedPlan

	fmt.Println("=== Task6 500万数据压测结果 ===")
	fmt.Printf("数据量: %d\n", *rowCount)
	printResult(baseline)
	printResult(optimized)
}

func setupSchema(db *sql.DB) {
	execMust(db, `
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  username TEXT NOT NULL,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  keyword_index TEXT NOT NULL,
  env TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
`)
}

func seedData(db *sql.DB, rowCount int) {
	start := time.Now()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("INSERT INTO users(id, username, name, email, keyword_index, env, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	envs := []string{"prod", "staging", "dev"}
	batchSize := 10000
	for i := 1; i <= rowCount; i++ {
		env := envs[i%len(envs)]
		username := fmt.Sprintf("user-%d", i)
		name := fmt.Sprintf("normal-user-%d", i)
		email := fmt.Sprintf("user-%d@example.com", i)

		if i%50000 == 0 {
			username = fmt.Sprintf("sre-ops-target-%d", i)
			name = fmt.Sprintf("sre ops target engineer %d", i)
			email = fmt.Sprintf("sre.ops.target.%d@example.com", i)
			env = "prod"
		}

		keywordIndex := strings.ToLower(username + " " + name + " " + email)
		if _, err := stmt.Exec(i, username, name, email, keywordIndex, env, i); err != nil {
			panic(err)
		}

		if i%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				panic(err)
			}
			tx, err = db.Begin()
			if err != nil {
				panic(err)
			}
			stmt, err = tx.Prepare("INSERT INTO users(id, username, name, email, keyword_index, env, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
			if err != nil {
				panic(err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	fmt.Printf("seed completed in %s\n", time.Since(start))
}

func explainQueryPlan(db *sql.DB, query string, args ...interface{}) []string {
	rows, err := db.Query("EXPLAIN QUERY PLAN "+query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	plan := make([]string, 0, 8)
	for rows.Next() {
		var id, parent, notUsed int
		var detail string
		if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
			panic(err)
		}
		plan = append(plan, detail)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return plan
}

func runBenchmark(db *sql.DB, label, query string, runs int, args ...interface{}) benchmarkResult {
	result := benchmarkResult{
		label:     label,
		durations: make([]time.Duration, 0, runs),
	}

	for i := 0; i < runs; i++ {
		start := time.Now()
		rows, err := db.Query(query, args...)
		if err != nil {
			panic(err)
		}
		rowCount := 0
		for rows.Next() {
			var id int64
			var username, name, email string
			if err := rows.Scan(&id, &username, &name, &email); err != nil {
				panic(err)
			}
			rowCount++
		}
		if err := rows.Close(); err != nil {
			panic(err)
		}
		duration := time.Since(start)
		result.durations = append(result.durations, duration)
		result.resultRows = rowCount
	}

	total := time.Duration(0)
	result.minimum = result.durations[0]
	for _, d := range result.durations {
		total += d
		if d > result.maximum {
			result.maximum = d
		}
		if d < result.minimum {
			result.minimum = d
		}
	}
	result.average = total / time.Duration(len(result.durations))
	return result
}

func printResult(result benchmarkResult) {
	fmt.Printf("\n[%s]\n", result.label)
	fmt.Printf("rows returned: %d\n", result.resultRows)
	fmt.Printf("durations: %s\n", joinDurations(result.durations))
	fmt.Printf("avg: %s, min: %s, max: %s\n", result.average, result.minimum, result.maximum)
	fmt.Println("plan:")
	for _, line := range result.plan {
		fmt.Printf("- %s\n", line)
	}
}

func joinDurations(durations []time.Duration) string {
	parts := make([]string, 0, len(durations))
	for _, d := range durations {
		parts = append(parts, d.String())
	}
	return strings.Join(parts, ", ")
}

func execMust(db *sql.DB, query string) {
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
}
