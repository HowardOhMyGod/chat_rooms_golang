package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "message"
    // "time"
    "os/signal"
    "syscall"
)

func input(title string) string {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print(title)
    str, _ := reader.ReadString('\n')

    return str
}

func showServerInput(conn net.Conn) {
    for {
        res, err := message.ReadText(conn)
        if err != nil {
            os.Exit(0)
        }
        fmt.Println(res)
        // fmt.Print("> ")
    }
}

func userInput(conn net.Conn) {
    for {
        str := input("")

        if str == "exit\n" {
            fmt.Print("bye")
            conn.Close()
            os.Exit(0)
        }
        message.SendText(conn, str)
    }
}

func sigHandler(sigChan chan os.Signal) {
    <- sigChan
    fmt.Print("bye")
    os.Exit(0)
}

func main() {
    // register SIGINT signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT)
    go sigHandler(sigChan)

    // initial connection
    conn, err := net.Dial("tcp", "127.0.0.1:8080")
    if err != nil {
        panic(err)
    }
    fmt.Println("Successfully connected!")
    name := input("Enter your name: ")

    // enter client's name
    message.SendText(conn, name)

    go showServerInput(conn)
    userInput(conn)
}
