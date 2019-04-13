package main

import (
    "fmt"
    "net"
    "os"
    "message"
    "github.com/satori/go.uuid"
    "errors"
    // "strings"
)

type Client struct {
    name string
    addr net.Addr
    conn net.Conn
}

type Room struct {
    id uuid.UUID
    clients map[net.Addr]Client
}


var options = "(1): Create room, (2) List rooms, (3) Join room, (4) Delete room"
var delChan = make(chan uuid.UUID)
var users = make(map[string]bool)

func isUniqName(name string) bool {
    if _, exist := users[name]; exist {
        return false
    }

    return true
}

func createRoom (rooms *map[uuid.UUID]*Room, client Client) uuid.UUID {
    room := new(Room)

    room.id = uuid.Must(uuid.NewV4())
    room.clients = make(map[net.Addr]Client)
    room.clients[client.addr] = client

    (*rooms)[room.id] = room

    return room.id
}

func joinRoom(rooms *map[uuid.UUID]*Room, roomId uuid.UUID, client Client) error {
    if _, exist := (*rooms)[roomId]; exist {
        (*rooms)[roomId].clients[client.addr] = client
        broadCast(rooms, roomId, client, fmt.Sprintf("%s join the room!", client.name))
        return nil
    } else {
        return errors.New("ID not exist.")
    }

}

func listRoom(rooms *map[uuid.UUID]*Room) string {
    // ID: oiefhioe-fehifef-24, Clients: Howard, Anna\n
    var roomList string

    if len(*rooms) == 0 {
        return "There is no room."
    }
    for uid, room := range *rooms {
        room_str := fmt.Sprintf("ID: %s, Clients: ", uid)
        for _, client := range room.clients {
            room_str += fmt.Sprintf("%s, ", client.name)
        }
        room_str += "\n"
        roomList += room_str
    }

    fmt.Println(roomList)
    return roomList
}

func leaveRoom(rooms *map[uuid.UUID]*Room, roomId uuid.UUID, client Client) {
    room := (*rooms)[roomId]
    delete((*room).clients, client.addr)
    broadCast(rooms, roomId, client, fmt.Sprintf("%s leaved the room!", client.name))

    // check if room is empty
    if len((*room).clients) == 0 {
        delete(*rooms, roomId)
    }
}

func delRoom(rooms *map[uuid.UUID]*Room, roomId string) bool {
    id, _ := uuid.FromString(roomId)

    if _, exist := (*rooms)[id]; exist {
        delete(*rooms, id)
        delChan <- id
        return true
    }

    return false
}

func broadCast(rooms *map[uuid.UUID]*Room, roomId uuid.UUID, client Client, text string) {
    room := (*rooms)[roomId]

    for addr, member := range room.clients {
        if addr != client.addr {
            message.SendText(member.conn, fmt.Sprintf("[%s]: %s", client.name, text))
        }
    }
}

func chat(conn net.Conn, rooms *map[uuid.UUID]*Room, roomId uuid.UUID, client Client) {
    message.SendText(conn, fmt.Sprintf("Enter room: %s", roomId))

    for {
        res, err := message.ReadText(conn)
        if err != nil { // client signal exit
            leaveRoom(rooms, roomId, client)
            return
        }

        // check if room has been deleted by another user
        select {
        case delRoomId:= <- delChan:
            if delRoomId == roomId {
                message.SendText(conn, "Someone delete this room. Back to lobby!")
                return
            }
        default:
        }

        if res == ":q" {
            leaveRoom(rooms, roomId, client)
            message.SendText(conn, "Back to lobby.")
            message.SendText(conn, options)
            return
        }
        // send message to members
        broadCast(rooms, roomId, client, res)
    }
}

func room_operation(conn net.Conn, rooms *map[uuid.UUID]*Room, client Client) error {
    message.SendText(conn, options)

    for {
        res, err := message.ReadText(conn)
        if err != nil {
            return err
        }

        // handler
        if res == "1" {
            roomId := createRoom(rooms, client)
            chat(conn, rooms, roomId, client)
        } else if res == "2" { // list room
            roomList := listRoom(rooms)
            message.SendText(conn, roomList)
        } else if res == "3" { // join room
            message.SendText(conn, "Enter room ID: ")
            res, err := message.ReadText(conn)

            if err != nil {
                break
            }

            room_id, _ := uuid.FromString(res)
            if err := joinRoom(rooms, room_id, client); err != nil {
                message.SendText(conn, "Room ID not found.")
            } else {
                chat(conn, rooms, room_id, client)
            }
        } else if res == "4" { //delete room
            message.SendText(conn, "Enter room ID you want to delete: ")
            id, err := message.ReadText(conn)

            if err != nil {
                break
            }

            success := delRoom(rooms, id)
            if !success {
                message.SendText(conn, "Room ID not found.")
            } else {
                message.SendText(conn, "Room deleted.")
            }

        } else {
            message.SendText(conn, options)
        }
    }

    return nil
}

func conn_handler(conn net.Conn, rooms *map[uuid.UUID]*Room) {
    var client Client
    var name string
    var err error

    for {
        name, err = message.ReadText(conn)
        if err != nil {
            return
        }

        if !isUniqName(name) {
            message.SendText(conn, "User name already exist.")
        } else {
            users[name] = true
            break
        }
    }

    client.name = name
    client.addr = conn.RemoteAddr()
    client.conn = conn

    fmt.Printf("Accept connection: %s, Name: %s\n", client.addr, client.name)

    for {
        if err := room_operation(conn, rooms, client); err != nil {
            delete(users, client.name)
            break
        }
    }
}

func main () {
    rooms := make(map[uuid.UUID]*Room)

    fmt.Println("Creating server...")
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Binding Error: ", err)
        os.Exit(1)
    }

    for {
        conn, _ := ln.Accept()
        go conn_handler(conn, &rooms)
    }
}
