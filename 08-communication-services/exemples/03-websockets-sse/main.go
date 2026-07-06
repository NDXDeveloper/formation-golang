/* ============================================================================
   Section 8.3 : WebSockets et Server-Sent Events
   Description : Les deux techniques de poussée serveur, auto-démonstratives
                 (serveurs httptest + clients intégrés). SSE : le handler de
                 la section (text/event-stream, ResponseController.Flush,
                 déconnexion via r.Context(), battement de cœur « : ping ») lu
                 comme le ferait EventSource. WebSocket (coder/websocket) :
                 Accept avec OriginPatterns — une origine étrangère est
                 REFUSÉE (403) —, écho JSON via wsjson, et la limite de
                 lecture (32 Kio par défaut → StatusMessageTooBig, parade
                 SetReadLimit).
   Fichier source : 03-websockets-sse.md
   Lancer : go run .        (aucun service requis)
   ============================================================================ */

package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// sseHandler — le handler de la section (ticker resserré pour la démo),
// battement de cœur inclus : une ligne « : … » est un commentaire ignoré
// par EventSource, qui maintient la connexion à travers les proxies.
func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w) // Go 1.20+ : plus net que w.(http.Flusher)
	ctx := r.Context()                  // réponse HTTP normale : r.Context() convient

	fmt.Fprint(w, ": ping\n\n") // battement de cœur (commentaire SSE)
	_ = rc.Flush()

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()

	for i := 1; ; i++ {
		select {
		case <-ctx.Done(): // le client s'est déconnecté
			return
		case <-ticker.C:
			fmt.Fprintf(w, "data: événement %d\n\n", i)
			if err := rc.Flush(); err != nil { // pousse immédiatement
				return
			}
		}
	}
}

// wsHandler — le serveur WebSocket de la section : Accept (origines en liste
// blanche), contexte DÉDIÉ (la connexion est « hijackée »), écho wsjson.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"example.com"}, // sinon le cross-origin est refusé
	})
	if err != nil {
		return
	}
	defer c.CloseNow()

	ctx := context.Background() // à dériver du contexte d'arrêt du serveur en pratique

	for {
		var msg any
		if err := wsjson.Read(ctx, c, &msg); err != nil {
			return // client parti ou fermeture
		}
		if err := wsjson.Write(ctx, c, msg); err != nil { // écho
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== SSE : lire le flux comme EventSource ===")
	sse := httptest.NewServer(http.HandlerFunc(sseHandler))
	defer sse.Close()

	cctx, stop := context.WithCancel(ctx)
	req, _ := http.NewRequestWithContext(cctx, "GET", sse.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("Content-Type :", resp.Header.Get("Content-Type"))
	sc := bufio.NewScanner(resp.Body)
	var data, comments int
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, ": "):
			comments++ // le battement de cœur
		case strings.HasPrefix(line, "data: "):
			data++
			fmt.Println("reçu →", line)
		}
		if data == 2 {
			break
		}
	}
	stop() // déconnexion : côté serveur, ctx.Done() fait sortir la boucle
	resp.Body.Close()
	fmt.Println("battements de cœur vus :", comments, "· événements :", data)

	fmt.Println()
	fmt.Println("=== WebSocket : origine refusée, puis écho JSON ===")
	ws := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer ws.Close()
	wsURL := "ws" + strings.TrimPrefix(ws.URL, "http")

	_, r2, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://evil.example.org"}},
	})
	fmt.Println("origine étrangère → refus :", err != nil, "· statut :", r2.StatusCode)

	c, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": []string{"https://example.com"}},
	})
	if err != nil {
		panic(err)
	}
	if err := wsjson.Write(ctx, c, map[string]any{"type": "ping", "n": 1}); err != nil {
		panic(err)
	}
	var echo map[string]any
	if err := wsjson.Read(ctx, c, &echo); err != nil {
		panic(err)
	}
	fmt.Println("écho reçu :", echo["type"], echo["n"])
	_ = c.Close(websocket.StatusNormalClosure, "fin")

	fmt.Println()
	fmt.Println("=== La limite de lecture : 32 Kio par défaut, parade SetReadLimit ===")
	lim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.CloseNow()
		// (défaut : 32 Kio par message ; décommentez pour accepter plus)
		// c.SetReadLimit(64 * 1024)
		for {
			if _, _, err := c.Read(context.Background()); err != nil {
				return
			}
		}
	}))
	defer lim.Close()
	c2, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(lim.URL, "http"), nil)
	if err != nil {
		panic(err)
	}
	_ = c2.Write(ctx, websocket.MessageBinary, make([]byte, 40*1024)) // 40 Kio > 32 Kio
	_, _, err = c2.Read(ctx)                                          // le serveur a fermé la connexion
	fmt.Println("message de 40 Kio → fermeture :", websocket.CloseStatus(err), "(la parade : SetReadLimit)")
	c2.CloseNow()
}
