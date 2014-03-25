package main

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

type BattleConnection struct {
    ws *websocket.Conn

    trainer *Trainer
}

func (bc *BattleConnection) reader() {
    for {
        _, msg, err := bc.ws.ReadMessage()
        if err != nil {
            bc.ws.WriteMessage(websocket.TextMessage, []byte("Error reading Message"))
            break
        }
        log.Println("Received", string(msg))

        if bc.trainer != nil && bc.trainer.battling {
            actionMessage := ActionMessage{Attack: -1, Switch: -1}
            _ = json.Unmarshal(msg, &actionMessage)
            actionMessage.trainer = bc.trainer
            bc.trainer.action <-&actionMessage
        } else {
            battle := PendingBattle{createdTime: time.Now(), conn: bc}
            _ = json.Unmarshal(msg, &battle)
            pendingBattles.toAdd <- &battle

            trainer := trainers[battle.Trainer_id]
            bc.trainer = trainer
            battle.trainer = trainer
            trainer.connections[bc] = true
        }
    }
}

