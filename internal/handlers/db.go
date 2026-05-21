package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"vickygo/internal/store"
)

// ---- helpers ----

func dbJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getDBSession(r *http.Request) (*store.DBSession, string, bool) {
	sid := r.Header.Get("X-DB-Session")
	if sid == "" {
		return nil, "", false
	}
	sess, ok := store.Global.GetDBSession(sid)
	return sess, sid, ok
}

// buildDSN constructs a driver-specific DSN from the connection form fields.
func buildDSN(driver, host, port, user, pass, dbname string) (string, error) {
	switch driver {
	case "mysql":
		if port == "" {
			port = "3306"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=10s",
			user, pass, host, port, dbname), nil
	case "postgres":
		if port == "" {
			port = "5432"
		}
		sslmode := "disable"
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
			host, port, user, pass, dbname, sslmode), nil
	case "sqlite":
		// host field is re-used as the file path for SQLite
		path := host
		if path == "" {
			path = ":memory:"
		}
		return path, nil
	}
	return "", fmt.Errorf("unsupported driver: %s", driver)
}

func quoteIdentifier(name string) string {
	if name == "" {
		return ""
	}
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// ---- DB Connect ----

func DBConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Driver string `json:"driver"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		User   string `json:"user"`
		Pass   string `json:"pass"`
		DBName string `json:"dbname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	driverName := req.Driver
	if driverName == "sqlite" {
		driverName = "sqlite" // modernc driver registers as "sqlite"
	}

	dsn, err := buildDSN(req.Driver, req.Host, req.Port, req.User, req.Pass, req.DBName)
	if err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		dbJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to open: " + err.Error()})
		return
	}

	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		dbJSON(w, http.StatusBadGateway, map[string]string{"error": "connection failed: " + err.Error()})
		return
	}

	sid := uuid.New().String()
	store.Global.SaveDBSession(sid, &store.DBSession{
		DB:     db,
		Driver: req.Driver,
		DSN:    dsn,
		DBName: req.DBName,
	})

	dbJSON(w, http.StatusOK, map[string]string{"session": sid, "driver": req.Driver, "dbname": req.DBName})
}

// ---- DB Disconnect ----

func DBDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	_, sid, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}
	store.Global.DeleteDBSession(sid)
	dbJSON(w, http.StatusOK, map[string]string{"ok": "disconnected"})
}

// ---- List Databases ----

func DBDatabasesHandler(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	var query string
	switch sess.Driver {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgres":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname"
	case "sqlite":
		dbJSON(w, http.StatusOK, map[string]any{"databases": []string{"main"}})
		return
	default:
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported driver"})
		return
	}

	rows, err := sess.DB.Query(query)
	if err != nil {
		dbJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			dbs = append(dbs, name)
		}
	}
	dbJSON(w, http.StatusOK, map[string]any{"databases": dbs})
}

// ---- List Tables ----

func DBTablesHandler(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	dbName := r.URL.Query().Get("db")

	var query string
	switch sess.Driver {
	case "mysql":
		if dbName == "" {
			dbName = sess.DBName
		}
		query = fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s' ORDER BY table_name", dbName)
	case "postgres":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name"
	case "sqlite":
		query = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
	default:
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported driver"})
		return
	}

	rows, err := sess.DB.Query(query)
	if err != nil {
		dbJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}
	if tables == nil {
		tables = []string{}
	}
	dbJSON(w, http.StatusOK, map[string]any{"tables": tables})
}

// ---- Database Objects ----

func DBObjectsHandler(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	dbName := r.URL.Query().Get("db")
	if dbName == "" {
		dbName = sess.DBName
	}

	var tables, views, routines []string
	switch sess.Driver {
	case "mysql":
		if dbName == "" {
			dbName = sess.DBName
		}
		tables = fetchStrings(sess.DB, "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE' ORDER BY table_name", dbName)
		views = fetchStrings(sess.DB, "SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'VIEW' ORDER BY table_name", dbName)
		routines = fetchStrings(sess.DB, "SELECT routine_name FROM information_schema.routines WHERE routine_schema = ? ORDER BY routine_name", dbName)
	case "postgres":
		tables = fetchStrings(sess.DB, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name")
		views = fetchStrings(sess.DB, "SELECT table_name FROM information_schema.views WHERE table_schema = 'public' ORDER BY table_name")
		routines = fetchStrings(sess.DB, "SELECT routine_name FROM information_schema.routines WHERE routine_schema = 'public' ORDER BY routine_name")
	case "sqlite":
		tables = fetchStrings(sess.DB, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
		views = fetchStrings(sess.DB, "SELECT name FROM sqlite_master WHERE type='view' ORDER BY name")
		routines = []string{}
	default:
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported driver"})
		return
	}

	dbJSON(w, http.StatusOK, map[string]any{
		"tables":     tables,
		"views":      views,
		"procedures": routines,
	})
}

func fetchStrings(db *sql.DB, query string, args ...any) []string {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			results = append(results, name)
		}
	}
	return results
}

// ---- Server Status ----

func DBStatusHandler(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	status := map[string]any{"driver": sess.Driver}
	switch sess.Driver {
	case "mysql":
		var version, currentDB, host string
		sess.DB.QueryRow("SELECT VERSION()").Scan(&version)
		sess.DB.QueryRow("SELECT DATABASE()").Scan(&currentDB)
		sess.DB.QueryRow("SELECT @@hostname").Scan(&host)
		status["version"] = version
		status["currentDB"] = currentDB
		status["host"] = host
	case "postgres":
		var version, currentDB, host string
		sess.DB.QueryRow("SELECT version()").Scan(&version)
		sess.DB.QueryRow("SELECT current_database()").Scan(&currentDB)
		sess.DB.QueryRow("SELECT inet_server_addr()::text").Scan(&host)
		status["version"] = version
		status["currentDB"] = currentDB
		status["host"] = host
	case "sqlite":
		var version string
		sess.DB.QueryRow("SELECT sqlite_version()").Scan(&version)
		status["version"] = version
		status["currentDB"] = sess.DBName
		status["host"] = "local"
	default:
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported driver"})
		return
	}

	dbJSON(w, http.StatusOK, status)
}

// ---- Row Update ----

func DBRowUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	var req struct {
		DB       string         `json:"db"`
		Table    string         `json:"table"`
		Original map[string]any `json:"original"`
		Changes  map[string]any `json:"changes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Table == "" {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "table is required"})
		return
	}
	if len(req.Changes) == 0 {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "no changes provided"})
		return
	}
	if req.DB == "" {
		req.DB = sess.DBName
	}

	setParts := []string{}
	whereParts := []string{}
	args := []any{}
	placeholder := func(i int) string {
		if sess.Driver == "postgres" {
			return fmt.Sprintf("$%d", i)
		}
		return "?"
	}
	argIndex := 1
	for col, value := range req.Changes {
		setParts = append(setParts, fmt.Sprintf("%s = %s", quoteIdentifier(col), placeholder(argIndex)))
		args = append(args, value)
		argIndex++
	}
	if len(setParts) == 0 {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "no valid columns to update"})
		return
	}
	for col, value := range req.Original {
		whereParts = append(whereParts, fmt.Sprintf("%s = %s", quoteIdentifier(col), placeholder(argIndex)))
		args = append(args, value)
		argIndex++
	}
	if len(whereParts) == 0 {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "no original row data provided"})
		return
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		quoteIdentifier(req.Table), strings.Join(setParts, ", "), strings.Join(whereParts, " AND "))
	if req.DB != "" && sess.Driver == "mysql" {
		query = fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s", quoteIdentifier(req.DB), quoteIdentifier(req.Table), strings.Join(setParts, ", "), strings.Join(whereParts, " AND "))
	}

	res, err := sess.DB.Exec(query, args...)
	if err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	dbJSON(w, http.StatusOK, map[string]any{"affected": affected})
}

// ---- Table Schema ----

func DBSchemaHandler(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	table := r.URL.Query().Get("table")
	if table == "" {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "table is required"})
		return
	}

	type Column struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Nullable string `json:"nullable"`
		Key      string `json:"key"`
		Default  string `json:"default"`
		Extra    string `json:"extra"`
	}

	var cols []Column

	switch sess.Driver {
	case "mysql":
		dbName := r.URL.Query().Get("db")
		if dbName == "" {
			dbName = sess.DBName
		}
		query := fmt.Sprintf("DESCRIBE %s", quoteIdentifier(table))
		if dbName != "" {
			query = fmt.Sprintf("DESCRIBE %s.%s", quoteIdentifier(dbName), quoteIdentifier(table))
		}
		rows, err := sess.DB.Query(query)
		if err != nil {
			dbJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var c Column
			var def, extra sql.NullString
			rows.Scan(&c.Name, &c.Type, &c.Nullable, &c.Key, &def, &extra)
			c.Default = def.String
			c.Extra = extra.String
			cols = append(cols, c)
		}

	case "postgres":
		q := `SELECT column_name, data_type, is_nullable, '', column_default, ''
		      FROM information_schema.columns
		      WHERE table_schema = 'public' AND table_name = $1
		      ORDER BY ordinal_position`
		rows, err := sess.DB.Query(q, table)
		if err != nil {
			dbJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var c Column
			var def sql.NullString
			rows.Scan(&c.Name, &c.Type, &c.Nullable, &c.Key, &def, &c.Extra)
			c.Default = def.String
			cols = append(cols, c)
		}

	case "sqlite":
		rows, err := sess.DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
		if err != nil {
			dbJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, typ string
			var notNull int
			var dflt sql.NullString
			var pk int
			rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk)
			nullable := "YES"
			if notNull == 1 {
				nullable = "NO"
			}
			key := ""
			if pk == 1 {
				key = "PRI"
			}
			cols = append(cols, Column{Name: name, Type: typ, Nullable: nullable, Key: key, Default: dflt.String})
		}
	}

	if cols == nil {
		cols = []Column{}
	}
	dbJSON(w, http.StatusOK, map[string]any{"columns": cols, "table": table})
}

// ---- Execute Query ----

func DBQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	sess, _, ok := getDBSession(r)
	if !ok {
		dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "no active session"})
		return
	}

	var req struct {
		SQL   string `json:"sql"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	q := strings.TrimSpace(req.SQL)
	if q == "" {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "sql is required"})
		return
	}
	if req.Limit <= 0 || req.Limit > 5000 {
		req.Limit = 500
	}

	start := time.Now()

	upper := strings.ToUpper(q)
	isSelect := strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "DESCRIBE") ||
		strings.HasPrefix(upper, "PRAGMA") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "WITH")

	if isSelect {
		rows, err := sess.DB.Query(q)
		if err != nil {
			dbJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		var results []map[string]any
		for rows.Next() {
			if len(results) >= req.Limit {
				break
			}
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			rows.Scan(ptrs...)
			row := make(map[string]any, len(cols))
			for i, col := range cols {
				v := vals[i]
				if b, ok := v.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = v
				}
			}
			results = append(results, row)
		}
		if results == nil {
			results = []map[string]any{}
		}

		dbJSON(w, http.StatusOK, map[string]any{
			"columns": cols,
			"rows":    results,
			"count":   len(results),
			"timeMs":  time.Since(start).Milliseconds(),
		})
		return
	}

	// Non-SELECT (INSERT/UPDATE/DELETE/CREATE/DROP …)
	res, err := sess.DB.Exec(q)
	if err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	dbJSON(w, http.StatusOK, map[string]any{
		"affected": affected,
		"timeMs":   time.Since(start).Milliseconds(),
		"message":  fmt.Sprintf("Query OK, %d row(s) affected", affected),
	})
}
