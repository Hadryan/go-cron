package main

import "os"
import "os/exec"
import "strings"
import "sync"
import "os/signal"
import "syscall"
import "github.com/robfig/cron"

import (
	"fmt"
	"net/http"
)

var last_err error

func execute(command string, args []string) {

	println("executing:", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	last_err = cmd.Run()
	cmd.Wait()
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	var schedule string = os.Args[1]
	var command string = os.Args[2]
	var args []string = os.Args[3:len(os.Args)]

	wg := &sync.WaitGroup{}

	c := cron.New()
	println("new cron:", schedule)

	c.AddFunc(schedule, func() {
		wg.Add(1)
		execute(command, args)
		wg.Done()
	})

	return c, wg
}

func handler(w http.ResponseWriter, r *http.Request) {
	if last_err != nil {
		http.Error(w, last_err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "OK")
}

func http_server(c *cron.Cron, wg *sync.WaitGroup) {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":18080", nil)
}

func start(c *cron.Cron, wg *sync.WaitGroup) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	println("Stopping")
	c.Stop()
	println("Waiting")
	wg.Wait()
	println("Exiting")
	os.Exit(0)
}

func main() {

	c, wg := create()

	go start(c, wg)
	go http_server(c, wg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	println(<-ch)
	stop(c, wg)
}
