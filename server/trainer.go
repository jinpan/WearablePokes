package main

import (
    "github.com/gorilla/websocket"
)

type Trainer struct {
    name string
    id string
    pokemon []Pokemon
    money uint64

    action chan *ActionMessage
    outbox chan []byte
    connections map[*BattleConnection] bool
    battling bool
}

func makeTrainer(id string) *Trainer {
    return trainers[id]
}

func (t *Trainer) run() {
    t.writer()
}

func (t *Trainer) writer() {
    for {
        msg := <-t.outbox
        for conn := range t.connections {
            err := conn.ws.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                conn.ws.Close()
            }
        }
    }
}

func (t *Trainer) isWiped() bool {
    for _, pokemon := range t.pokemon {
        if pokemon.state.health > 0{
            return false
        }
    }
    return true
}

