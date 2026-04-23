package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robinv8/mino-runtime/internal/brief"
	"github.com/robinv8/mino-runtime/internal/event"
	"github.com/robinv8/mino-runtime/internal/lock"
	"github.com/robinv8/mino-runtime/internal/state"
)

//go:embed frontend
var frontendFS embed.FS

// Server is the Phase 2 HTTP/WebSocket runtime.
type Server struct {
	repoRoot  string
	addr      string
	upgrader  websocket.Upgrader
	clients   map[*client]bool
	clientsMu sync.RWMutex
	broadcast chan Event
	commands  chan CommandRequest
	httpSrv   *http.Server
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

// Event is pushed to all connected WebSocket clients.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

type client struct {
	conn   *websocket.Conn
	send   chan Event
	server *Server
}

// New creates a Server bound to the given repo root.
func New(repoRoot, addr string) *Server {
	s := &Server{
		repoRoot: repoRoot,
		addr:     addr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // local dev
		},
		clients:   make(map[*client]bool),
		broadcast: make(chan Event, 64),
		commands:  make(chan CommandRequest, 16),
	}
	return s
}

// Start runs the HTTP server and command processor.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/commands", s.handleCommands)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/state", s.handleState)

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
	// TODO: scan .mino/briefs/ and return all task states
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": []interface{}{},
	})
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

// --- WebSocket ---

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &client{
		conn:   conn,
		send:   make(chan Event, 16),
		server: s,
	}

	s.clientsMu.Lock()
	s.clients[c] = true
	s.clientsMu.Unlock()

	go c.writePump()
	go c.readPump()

	// Send immediate heartbeat
	c.send <- Event{Type: "connected", Timestamp: time.Now().UTC()}
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

func (s *Server) emit(ev Event) {
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
		s.emit(Event{
			Type:      "command_completed",
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
