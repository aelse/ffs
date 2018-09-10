package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	"github.com/aelse/ffs"
)

func setupSignalHandler(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalHandler(cancel)

	go runWebApp(ctx)

	<-ctx.Done()
	fmt.Println("Waiting before shutdown")
	time.Sleep(time.Second)
	fmt.Println("Shutting down")
}

func runWebApp(ctx context.Context) {
	r := gin.Default()
	m := melody.New()

	var flags = make(map[string]ffs.FeatureFlag, 0)

	r.GET("/", func(c *gin.Context) {
		c.Writer.Write([]byte("ffs is running"))
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	r.POST("/flags", func(c *gin.Context) {
		var f ffs.FeatureFlag
		if err := c.BindJSON(&f); err != nil || f.Name == "" {
			log.Printf("Invalid flag data: %v", err)
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		flags[f.Name] = f
		b, err := json.Marshal(f)
		if err != nil {
			log.Printf("Could not marshal flag: %v", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		m.Broadcast(b)
		c.Writer.WriteHeader(http.StatusNoContent)
	})

	sendFlags := func(s *melody.Session) {
		for _, f := range flags {
			b, err := json.Marshal(f)
			if err != nil {
				log.Printf("Could not marshal flag: %v", err)
				continue
			}
			s.Write(b)
		}
	}

	m.HandleConnect(func(s *melody.Session) {
		// Client connected, send flags.
		sendFlags(s)
	})

	m.HandleMessage(func(s *melody.Session, _ []byte) {
		// Client said something, send flags.
		sendFlags(s)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		log.Printf("Shutting down web server")
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		server.Shutdown(ctx)
	}()

	fmt.Printf("Listening on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Web server terminated: %v", err)
	}
}
