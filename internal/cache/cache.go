package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type Store struct {
	db     *sql.DB
	root   string
	limitB int64
	log    zerolog.Logger
}

func New(dbPath string, limitGB int, log zerolog.Logger) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS cache (
		key TEXT PRIMARY KEY,
		path TEXT,
		size INTEGER,
		etag TEXT,
		last_modified TEXT,
		expire_at INTEGER
	)`)

	return &Store{
		db:     db,
		root:   filepath.Dir(dbPath),
		limitB: int64(limitGB) * 1_000_000_000,
		log:    log,
	}, nil
}

func (s *Store) Close() { _ = s.db.Close() }

func hashKey(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

// Lookup returns local path if cache valid, otherwise empty string.
func (s *Store) Lookup(req *http.Request) (localPath, etag, modTime string, ok bool) {
	key := hashKey(req.URL.String())
	row := s.db.QueryRow(`SELECT path, etag, last_modified, expire_at FROM cache WHERE key=?`, key)
	var p, e, m string
	var exp int64
	if err := row.Scan(&p, &e, &m, &exp); err != nil {
		return
	}
	if time.Now().Unix() > exp {
		return
	}
	return p, e, m, true
}

func (s *Store) Save(resp *http.Response, body []byte, ttl time.Duration) (string, error) {
	key := hashKey(resp.Request.URL.String())
	filename := filepath.Join(s.root, key)
	if err := os.WriteFile(filename, body, 0o644); err != nil {
		return "", err
	}
	_, err := s.db.Exec(`INSERT OR REPLACE INTO cache(key, path, size, etag, last_modified, expire_at)
		VALUES(?,?,?,?,?,?)`,
		key, filename, len(body),
		resp.Header.Get("Etag"),
		resp.Header.Get("Last-Modified"),
		time.Now().Add(ttl).Unix(),
	)
	return filename, err
}
