package message

import (
    "bytes"
    "encoding/gob"
    "net"
    "strings"
    "fmt"
    "errors"
    // "fmt"
)

type Message struct {
    Text string
}

func Encode(msg *Message) []byte {
    var msg_bytes bytes.Buffer
    e := gob.NewEncoder(&msg_bytes)

    if err := e.Encode(msg); err != nil {
        panic(err)
    }

    return msg_bytes.Bytes()
}

func Decode(tmp []byte) *Message {
    msg_bytes := bytes.NewBuffer(tmp)

    msg := new(Message)
    d := gob.NewDecoder(msg_bytes)

    if err := d.Decode(msg); err != nil {
        panic(err)
    }

    return msg
}

func SendMsg(conn net.Conn, msg *Message) {
    msg_bytes := Encode(msg)
    conn.Write(msg_bytes)
}

func ReadMsg(conn net.Conn) (*Message, error) {
    var msg *Message
    tmp := make([]byte, 2048)

    _, err := conn.Read(tmp)
    if err != nil {
        fmt.Printf("%s closed the connection.\n", conn.RemoteAddr())
        return nil, errors.New("Client closed the connection.")

    }

    msg = Decode(tmp)
    return msg, nil
}

func SendText(conn net.Conn, text string) {
    msg := new(Message)
    trim := strings.TrimSuffix(text, "\n")
    msg.Text = trim

    SendMsg(conn, msg)
}

func ReadText(conn net.Conn) (string, error) {
    msg, err := ReadMsg(conn)
    if err != nil {
        return "", err
    }

    return strings.TrimSuffix(string(msg.Text), "\n"), err
}
