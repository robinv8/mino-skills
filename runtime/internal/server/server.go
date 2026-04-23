package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robinv8/mino-runtime/internal/brief"
	"github.com/robinv8/mino-runtime/internal/event"
	"github.com/robinv8/mino-runtime/internal/lock"
	"github.com/robinv8/mino-runtime/internal/state"
	"github.com/robinv8/mino-runtime/pkg/schema"
)

//go:embed frontend
var frontendFS embed.FS

const version = "0.1.0"

// Server is the Phase 2 HTTP/WebSocket runtime.
type Server struct {
	repoRoot  string
	addr      string
	upgrader  websocket.Upgrader
	clients   map[*client]bool
	clientsMu sync.RWMutex
	broadcast chan schema.Event
	commands  chan CommandRequest
	httpSrv   *http.Server
	startedAt time.Time
	token     string
}

// CommandRequest is what clients POST to /commands.
type CommandRequest struct {
	ID      string                 `json:"command_id"`
	Type    string                 `json:"type"`
	Target  map[string]interface{} `json:"target,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	DryRun  bool                   `json:"dry_run,omitempty"`
}

// CommandResponse is the 202 Accepted body.
type CommandResponse struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"`
}

type client struct {
	conn   *websocket.Conn
	send   chan schema.Event
	server *Server
}

// New creates a Server bound to the given repo root.
func New(repoRoot, addr string) *Server {
	s := &Server{
		repoRoot:  repoRoot,
		addr:      addr,
		startedAt: time.Now().UTC(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // local dev
		},
		clients:   make(map[*client]bool),
		broadcast: make(chan schema.Event, 64),
		commands:  make(chan CommandRequest, 16),
	}
	s.loadToken()
	return s
}

func (s *Server) loadToken() {
	// Prefer environment variable
	if t := os.Getenv("MINO_DAEMON_TOKEN"); t != "" {
		s.token = t
		return
	}
	// Fallback to token file
	path := filepath.Join(s.repoRoot, ".mino", "daemon.token")
	data, err := os.ReadFile(path)
	if err == nil {
		s.token = strings.TrimSpace(string(data))
	}
}

// Start runs the HTTP server and command processor.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Auth wrapper for API routes
	api := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !s.requireAuth(w, r) {
				return
			}
			h(w, r)
		}
	}

	// API routes (v1)
	mux.HandleFunc("/api/v1/commands", api(s.handleCommands))
	mux.HandleFunc("/api/v1/ws", s.handleWebSocket)
	mux.HandleFunc("/api/v1/tasks", api(s.handleTasks))
	mux.HandleFunc("/api/v1/state", api(s.handleState))
	mux.HandleFunc("/api/v1/events", api(s.handleEvents))
	mux.HandleFunc("/api/v1/health", api(s.handleHealth))

	// Static frontend — serve frontend/ dir contents at root
	fsRoot, _ := fs.Sub(frontendFS, "frontend")
	mux.Handle("/", http.FileServer(http.FS(fsRoot)))

	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go s.commandProcessor()
	go s.broadcaster()

	fmt.Printf("[serve] Listening on http://%s\n", s.addr)
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// --- HTTP handlers ---

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.ID = fmt.Sprintf("cmd-%d", time.Now().UnixNano())
	select {
	case s.commands <- req:
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(CommandResponse{
			CommandID: req.ID,
			Status:    "accepted",
		})
	default:
		http.Error(w, "command queue full", http.StatusServiceUnavailable)
	}
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	briefsDir := filepath.Join(s.repoRoot, ".mino", "briefs")
	entries, err := os.ReadDir(briefsDir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"tasks": []interface{}{}})
		return
	}

	includeEvents := r.URL.Query().Get("include") == "events"
	var tasks []map[string]interface{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "issue-") || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(briefsDir, entry.Name()))
		if err != nil {
			continue
		}
		b := brief.Parse(string(data))
		if b.IssueNumber == 0 {
			continue
		}
		task := map[string]interface{}{
			"issue":         b.IssueNumber,
			"task_key":      b.TaskKey,
			"stage":         b.CurrentStage,
			"next_stage":    b.NextStage,
			"attempt":       b.AttemptCount,
			"max_retry":     b.MaxRetryCount,
			"spec_revision": b.SpecRevision,
		}
		if includeEvents {
			eventsDir := filepath.Join(s.repoRoot, ".mino", "events", fmt.Sprintf("issue-%d", b.IssueNumber))
			task["events"] = s.scanEvents(eventsDir, 20)
		}
		tasks = append(tasks, task)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	issueFloat, ok := parseIssueFromQuery(r)
	if !ok {
		http.Error(w, "missing issue param", http.StatusBadRequest)
		return
	}
	issue := int(issueFloat)

	b, err := brief.Load(s.repoRoot, issue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"issue":        b.IssueNumber,
		"task_key":     b.TaskKey,
		"stage":        b.CurrentStage,
		"next_stage":   b.NextStage,
		"attempt":      b.AttemptCount,
		"max_retry":    b.MaxRetryCount,
		"spec_revision": b.SpecRevision,
	})
}

func parseIssueFromQuery(r *http.Request) (float64, bool) {
	issue := r.URL.Query().Get("issue")
	if issue == "" {
		return 0, false
	}
	var f float64
	fmt.Sscanf(issue, "%f", &f)
	return f, f > 0
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	issueFloat, ok := parseIssueFromQuery(r)
	if !ok {
		http.Error(w, "missing issue param", http.StatusBadRequest)
		return
	}
	issue := int(issueFloat)

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	since := 0
	if s := r.URL.Query().Get("since"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			since = v
		}
	}

	eventsDir := filepath.Join(s.repoRoot, ".mino", "events", fmt.Sprintf("issue-%d", issue))
	all := s.scanEvents(eventsDir, 0) // 0 = no limit

	// Filter and paginate
	var filtered []map[string]interface{}
	for _, ev := range all {
		seq, _ := ev["seq"].(int)
		if seq > since {
			filtered = append(filtered, ev)
		}
	}
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	nextSeq := since
	if len(filtered) > 0 {
		last := filtered[len(filtered)-1]
		if s, ok := last["seq"].(int); ok {
			nextSeq = s
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":  filtered,
		"next_seq": nextSeq,
	})
}

// scanEvents reads event files from the given directory and returns them sorted by seq.
// max = 0 means no limit.
func (s *Server) scanEvents(eventsDir string, max int) []map[string]interface{} {
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		return nil
	}

	type item struct {
		seq  int
		name string
	}
	var items []item
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Format: {seq:04d}-{event_type}.yml
		parts := strings.SplitN(name, "-", 2)
		if len(parts) != 2 {
			continue
		}
		seq, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		items = append(items, item{seq: seq, name: name})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].seq < items[j].seq
	})

	if max > 0 && len(items) > max {
		items = items[len(items)-max:] // take most recent
	}

	var result []map[string]interface{}
	for _, it := range items {
		data, err := os.ReadFile(filepath.Join(eventsDir, it.name))
		if err != nil {
			continue
		}
		ev := parseEventFile(it.seq, string(data))
		if ev != nil {
			result = append(result, ev)
		}
	}
	return result
}

// parseEventFile extracts fields from an event YAML block.
func parseEventFile(seq int, raw string) map[string]interface{} {
	// Extract YAML block between ```yaml and ```
	start := strings.Index(raw, "```yaml\n")
	if start == -1 {
		return nil
	}
	start += len("```yaml\n")
	end := strings.Index(raw[start:], "\n```")
	if end == -1 {
		return nil
	}
	yamlBlock := raw[start : start+end]

	var eventType, timestamp string
	payload := make(map[string]interface{})

	lines := strings.Split(yamlBlock, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "iron_tree:" {
			continue
		}
		// Look for "key: value" pairs at indent level 2
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			switch key {
			case "event_type":
				eventType = val
			case "timestamp":
				timestamp = val
			case "version", "sequence":
				// skip internal fields
			default:
				payload[key] = val
			}
		}
	}

	if eventType == "" {
		return nil
	}

	return map[string]interface{}{
		"seq":       seq,
		"type":      eventType,
		"timestamp": timestamp,
		"payload":   payload,
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":    version,
		"repo_root":  s.repoRoot,
		"pid":        os.Getpid(),
		"started_at": s.startedAt.Format(time.RFC3339),
	})
}

// requireAuth checks the Authorization header against the daemon token.
// Returns true if authenticated (or no token configured — backward compat).
func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if s.token == "" {
		return true // no token configured; allow
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) || auth[len(prefix):] != s.token {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// --- WebSocket ---

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &client{
		conn:   conn,
		send:   make(chan schema.Event, 16),
		server: s,
	}

	s.clientsMu.Lock()
	s.clients[c] = true
	s.clientsMu.Unlock()

	go c.writePump()
	go c.readPump()

	// Send immediate heartbeat
	c.send <- schema.Event{Type: schema.EventConnected, Timestamp: time.Now().UTC()}
}

func (s *Server) broadcaster() {
	for ev := range s.broadcast {
		s.clientsMu.RLock()
		for c := range s.clients {
			select {
			case c.send <- ev:
			default:
				// Client slow; drop event
			}
		}
		s.clientsMu.RUnlock()
	}
}

func (s *Server) emit(ev schema.Event) {
	select {
	case s.broadcast <- ev:
	default:
		// Broadcast queue full; drop
	}
}

func (c *client) readPump() {
	defer func() {
		c.server.clientsMu.Lock()
		delete(c.server.clients, c)
		c.server.clientsMu.Unlock()
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case ev, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteJSON(ev); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// --- Command processor ---

func (s *Server) commandProcessor() {
	for cmd := range s.commands {
		result := s.executeCommand(cmd)
		s.emit(schema.Event{
			Type:      schema.EventCommandCompleted,
			Timestamp: time.Now().UTC(),
			Payload: map[string]interface{}{
				"command_id": cmd.ID,
				"type":       cmd.Type,
				"result":     result,
			},
		})
	}
}

func (s *Server) executeCommand(cmd CommandRequest) map[string]interface{} {
	result := map[string]interface{}{"success": false}

	switch cmd.Type {
	case "step", "approve", "skip", "cancel", "retry":
		issue, _ := cmd.Target["issue"].(float64)
		if issue == 0 {
			result["error"] = "missing target.issue"
			return result
		}

		if cmd.DryRun {
			b, err := brief.Load(s.repoRoot, int(issue))
			if err != nil {
				result["error"] = err.Error()
				return result
			}
			from := state.Stage(b.CurrentStage)
			next, _ := state.DefaultNext(from)
			result["dry_run"] = true
			result["would_advance"] = fmt.Sprintf("%s → %s", from, next)
			result["success"] = true
			return result
		}

		if err := lock.Acquire(s.repoRoot, fmt.Sprintf("api-%s-issue-%d", cmd.Type, int(issue))); err != nil {
			result["error"] = err.Error()
			return result
		}
		defer lock.Release(s.repoRoot)

		b, err := brief.Load(s.repoRoot, int(issue))
		if err != nil {
			result["error"] = err.Error()
			return result
		}

		from := state.Stage(b.CurrentStage)
		var next state.Stage
		var evType string

		switch cmd.Type {
		case "step", "approve":
			next, _ = state.DefaultNext(from)
			evType = "task_advanced"
		case "skip":
			next, _ = state.DefaultNext(from)
			evType = "task_skipped"
		case "cancel":
			next = state.StageHalted
			evType = "task_cancelled"
		case "retry":
			if from == state.StageHalted {
				next = state.StageRun
			} else {
				result["error"] = "can only retry from halted state"
				return result
			}
			evType = "task_retried"
		}

		if next == "" {
			result["error"] = fmt.Sprintf("cannot %s from %s", cmd.Type, from)
			return result
		}

		updated := b.Patch("Current Stage", b.CurrentStage, string(next))
		nextNext, _ := state.DefaultNext(next)
		if nextNext != "" {
			updated = updated.Patch("Next Stage", b.NextStage, string(nextNext))
		}
		if from == state.StageTask {
			updated = updated.Patch("Attempt Count", fmt.Sprintf("%d", updated.AttemptCount), fmt.Sprintf("%d", updated.AttemptCount+1))
		}

		if err := updated.Save(s.repoRoot); err != nil {
			result["error"] = err.Error()
			return result
		}

		_ = event.Record(s.repoRoot, int(issue), evType, map[string]string{
			"from": string(from),
			"to":   string(next),
			"action": cmd.Type,
		})

		result["success"] = true
		result["advanced"] = fmt.Sprintf("%s → %s", from, next)
		result["action"] = cmd.Type

	default:
		result["error"] = fmt.Sprintf("unknown command type: %s", cmd.Type)
	}

	return result
}
