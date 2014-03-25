package main

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

type BattleConnection struct {
    ws *websocket.Conn

    send chan []byte
    action chan ActionMessage

    trainer *Trainer
    started bool
}

func (bc *BattleConnection) reader() {
    for {
        _, msg, err := bc.ws.ReadMessage()
        if err != nil {
            bc.ws.WriteMessage(websocket.TextMessage, []byte("Error reading Message"))
            break
        }
        log.Println("Received", string(msg))
        log.Println("Started?", bc.started)

        if bc.started {
            actionMessage := ActionMessage{Attack: -1, Switch: -1}
            _ = json.Unmarshal(msg, &actionMessage)
            actionMessage.trainer = bc.trainer
            bc.action <-actionMessage
            log.Println("put message onto channel for", bc.trainer.name)
        } else {
            battle := PendingBattle{createdTime: time.Now(), conn: bc}
            _ = json.Unmarshal(msg, &battle)
            pendingBattles.toAdd <- &battle
        }
    }
}

func (bc *BattleConnection) writer() {
    for {
        msg := <-bc.send
        log.Println("To send", string(msg))

        err := bc.ws.WriteMessage(websocket.TextMessage, msg)
        if err != nil {
            log.Fatal(err)
            break
        }
    }
    bc.ws.Close()
    log.Println("Writer exiting")
}
