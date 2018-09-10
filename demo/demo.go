package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aelse/ffs"
	"github.com/gin-gonic/gin"
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

	flagClient, err := ffs.NewClient("localhost:8080")
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	go runWebApp(ctx, flagClient)

	<-ctx.Done()
	fmt.Println("Waiting before shutdown")
	time.Sleep(time.Second)
	fmt.Println("Shutting down")
}

func runWebApp(ctx context.Context, flagClient *ffs.Client) {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Writer.Write([]byte("<html><body><p><ul>"))
		v := flagClient.Bool("myflag", false)
		if v {
			c.Writer.Write([]byte("<li>The flag is on"))
		} else {
			c.Writer.Write([]byte("<li>The flag is off"))
		}
		c.Writer.Write([]byte("</ul></p></body><html>"))
	})

	server := http.Server{
		Addr:    ":8081",
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
