package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed static/index.html
var staticFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSMessage represents a WebSocket JSON message
type WSMessage struct {
	Type    string `json:"type"`
	CmdName string `json:"cmdName,omitempty"`
	ExecId  int64  `json:"execId,omitempty"`
}

// WSConn wraps a WebSocket connection
type WSConn struct {
	id      string
	conn    *websocket.Conn
	manager *CmdManager
	mu      sync.Mutex
	done    chan struct{}
	closed  bool
}

func main() {
	dir := flag.String("dir", "", "包含 .cmd 文件的目录（必填）")
	port := flag.Int("port", 8080, "HTTP 服务端口")
	maxHistory := flag.Int("max-history", 5, "每个命令保留的最大历史记录数")
	flag.Parse()

	if *dir == "" {
		log.Fatal("必须指定 -dir 参数")
	}

	manager := NewCmdManager(*dir, *maxHistory)
	if err := manager.Scan(); err != nil {
		log.Fatalf("扫描目录失败: %v", err)
	}
	log.Printf("发现 %d 个 .cmd 文件", len(manager.cmdFiles))

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/ws", handleWebSocket(manager))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("服务启动在 %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "页面加载失败", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleWebSocket(manager *CmdManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}
		wsc := &WSConn{
			id:      generateID(),
			conn:    conn,
			manager: manager,
			done:    make(chan struct{}),
		}

		// Send initial cmdList
		wsc.sendCmdList()

		go wsc.writePump()
		go wsc.readPump()
	}
}

func (wsc *WSConn) readPump() {
	defer wsc.close()
	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			wsc.sendError("INVALID_MESSAGE", "消息格式错误")
			continue
		}
		switch msg.Type {
		case "run":
			wsc.handleRun(msg)
		case "watch":
			wsc.handleWatch(msg)
		case "replay":
			wsc.handleReplay(msg)
		case "history":
			wsc.handleHistory(msg)
		case "unwatch":
			wsc.handleUnwatch(msg)
		default:
			wsc.sendError("UNKNOWN_TYPE", fmt.Sprintf("未知的消息类型: %s", msg.Type))
		}
	}
}

func (wsc *WSConn) writePump() {
	defer wsc.close()
	for {
		select {
		case <-wsc.done:
			return
		}
	}
}

func (wsc *WSConn) handleRun(msg WSMessage) {
	record, err := wsc.manager.Start(msg.CmdName)
	if err != nil {
		wsc.sendError(parseErrorCode(err), err.Error())
		return
	}
	wsc.sendStatus(record)
	// Auto-watch after run
	wsc.handleWatch(msg)
}

func (wsc *WSConn) handleWatch(msg WSMessage) {
	sub, err := wsc.manager.Subscribe(msg.CmdName, wsc.id)
	if err != nil {
		wsc.sendError(parseErrorCode(err), err.Error())
		return
	}
	// Start forwarding output to WebSocket
	go wsc.forwardOutput(msg.CmdName, sub)
}

func (wsc *WSConn) handleReplay(msg WSMessage) {
	record, err := wsc.manager.FindRecord(msg.CmdName, msg.ExecId)
	if err != nil {
		wsc.sendError(parseErrorCode(err), err.Error())
		return
	}
	lines := record.Output.Replay(0)
	batchSize := 100
	for i := 0; i < len(lines); i += batchSize {
		end := i + batchSize
		if end > len(lines) {
			end = len(lines)
		}
		batch := lines[i:end]
		outMsg := map[string]interface{}{
			"type":       "replayOutput",
			"cmdName":    msg.CmdName,
			"execId":     msg.ExecId,
			"lines":      batch,
			"isComplete": end >= len(lines),
		}
		wsc.sendJSON(outMsg)
	}
}

func (wsc *WSConn) handleHistory(msg WSMessage) {
	records, err := wsc.manager.GetHistory(msg.CmdName)
	if err != nil {
		wsc.sendError(parseErrorCode(err), err.Error())
		return
	}
	histMsg := map[string]interface{}{
		"type":    "history",
		"cmdName": msg.CmdName,
		"records": records,
	}
	wsc.sendJSON(histMsg)
}

func (wsc *WSConn) handleUnwatch(msg WSMessage) {
	wsc.manager.Unsubscribe(msg.CmdName, wsc.id)
}

func (wsc *WSConn) forwardOutput(cmdName string, sub *Subscriber) {
	lock := wsc.manager.locks[cmdName]
	defer wsc.manager.Unsubscribe(cmdName, wsc.id)

	for line := range sub.Ch {
		lock.mu.Lock()
		var execId int64
		if lock.running != nil {
			execId = lock.running.ID
		}
		lock.mu.Unlock()

		outMsg := map[string]interface{}{
			"type":    "output",
			"cmdName": cmdName,
			"execId":  execId,
			"text":    line.Text,
			"stream":  line.Stream,
			"seq":     line.Seq,
		}
		wsc.sendJSON(outMsg)
	}
	// Channel closed = process ended, send final status
	wsc.sendCmdList() // refresh command states
}

func (wsc *WSConn) sendCmdList() {
	cmdList := wsc.manager.GetCmdList()
	wsc.sendJSON(map[string]interface{}{
		"type":      "cmdList",
		"commands":  cmdList,
	})
}

func (wsc *WSConn) sendStatus(record *ExecRecord) {
	msg := map[string]interface{}{
		"type":      "status",
		"cmdName":   record.CmdName,
		"execId":    record.ID,
		"status":    record.Status,
		"startTime": record.StartTime.Format("2006-01-02T15:04:05"),
	}
	if record.EndTime != nil {
		msg["endTime"] = record.EndTime.Format("2006-01-02T15:04:05")
		msg["exitCode"] = record.ExitCode
	} else {
		msg["endTime"] = nil
		msg["exitCode"] = -1
	}
	wsc.sendJSON(msg)
}

func (wsc *WSConn) sendError(code, message string) {
	wsc.sendJSON(map[string]interface{}{
		"type":    "error",
		"code":    code,
		"message": message,
	})
}

func (wsc *WSConn) sendJSON(v interface{}) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	wsc.conn.WriteMessage(websocket.TextMessage, data)
}

func (wsc *WSConn) close() {
	wsc.mu.Lock()
	if wsc.closed {
		wsc.mu.Unlock()
		return
	}
	wsc.closed = true
	wsc.mu.Unlock()
	wsc.conn.Close()
	close(wsc.done)
}

func parseErrorCode(err error) string {
	msg := err.Error()
	if len(msg) >= 10 && msg[:10] == "CMD_LOCKED" {
		return "CMD_LOCKED"
	}
	if len(msg) >= 13 && msg[:13] == "CMD_NOT_FOUND" {
		return "CMD_NOT_FOUND"
	}
	if len(msg) >= 16 && msg[:16] == "REPLAY_NOT_FOUND" {
		return "REPLAY_NOT_FOUND"
	}
	if len(msg) >= 15 && msg[:15] == "CMD_NOT_RUNNING" {
		return "CMD_NOT_RUNNING"
	}
	return "INTERNAL_ERROR"
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}