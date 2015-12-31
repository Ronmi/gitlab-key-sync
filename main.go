package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/inotify"
)

var (
	fw   bool
	port int
	user string
	gitlabKeyFile string = "gitlab.keys"
)

func init() {
	flag.BoolVar(&fw, "forward", false, "forward the request")
	flag.IntVar(&port, "port", 0, "gitlab ssh port")
	flag.StringVar(&user, "user", "git", "gitlab ssh user")
	flag.Parse()

	if flag.NArg() == 1 {
		gitlabKeyFile = flag.Arg(0)
	}

	if port <= 0 {
		log.Fatalf("You must specify a valid port.")
	}
}

func main() {
	if fw {
		forward()
		return
	}
	
	if _, err := os.Stat(gitlabKeyFile); err == nil {
		mv(convert(gitlabKeyFile))
	}
	
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatalf("Cannot initialize inotify: %s", err)
	}

	if err = watcher.AddWatch(gitlabKeyFile, inotify.IN_CREATE|inotify.IN_MODIFY); err != nil {
		log.Fatalf("Cannot watch %s: %s", gitlabKeyFile, err)
	}

	for {
		select {
		case <-watcher.Event:
			mv(convert(gitlabKeyFile))
		case err := <-watcher.Error:
			log.Printf("Got error: %s", err)
		}
	}
}

func convert(orig string) (err error) {
	in, err := os.Open(orig)
	if err != nil {
		log.Printf("Cannot open %s for read: %s", orig, err)
		return
	}
	defer in.Close()

	scanner := bufio.NewScanner(in)

	out, err := os.Create("authorized_keys.new")
	if err != nil {
		log.Printf("Cannot create new key file: %s", err)
		return
	}
	defer out.Close()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || ([]byte(line))[0] == '#' {
			continue
		}

		parts := strings.Split(line, "ssh-")
		parts[0] = fmt.Sprintf(`command="%s -forward -port %d" `, os.Args[0], port)
		_, err = fmt.Fprintln(out, strings.Join(parts, "ssh-"))
		if err != nil {
			log.Printf("Cannot write new key file: %s", err)
			return
		}
	}
	return
}

func mv(err error) {
	if err != nil {
		return
	}

	if err := os.Rename("authorized_keys.new", "authorized_keys"); err != nil {
		log.Printf("Cannot rename newly created key file to real key file: %s", err)
	}
}

func forward() {
	origCmd := os.Getenv("SSH_ORIGINAL_COMMAND")
	cmd := exec.Command(
		"sh", "-c",
		fmt.Sprintf("ssh -l git -p %d 127.0.0.1 %s", port, origCmd),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Printf("Error forwarding user session: %s", err)
	}
}
