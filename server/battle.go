package main

import (
    "log"
    "math"
    "math/rand"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

type PendingBattle struct {
    Trainer_id string `json:"trainer"`
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`

    trainer *Trainer
    createdTime time.Time
    conn *BattleConnection
}

type PendingBattles struct {
    data map[*PendingBattle] bool

    toAdd chan *PendingBattle
    toRemove chan *PendingBattle
}

type Battle struct {
    conn1 *BattleConnection
    conn2 *BattleConnection
}

func (battle1 *PendingBattle) CloseTo(battle2 *PendingBattle) bool {
    threshold := float64(10)

    R := float64(6367500)
    dlat := (battle2.Lat - battle1.Lat) * math.Pi / 180
    dlng := (battle2.Lng - battle1.Lng) * math.Pi / 180
    lat1 := battle1.Lat * math.Pi / 180
    lat2 := battle2.Lat * math.Pi / 180

    a := math.Sin(dlat/2) * math.Sin(dlat/2) * math.Sin(dlng/2) * math.Sin(dlng/2) * math.Cos(lat1) * math.Cos(lat2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    d := R * c

    if d < threshold {
        return true
    } else {
        return false
    }
}

func (pb *PendingBattle) remove() {
    pb.conn.ws.Close()
}

func battleHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if err != nil {
        return
    }
    log.Println("New Connection")

    bc := &BattleConnection{ws: ws}
    bc.reader()
}

func (pbs *PendingBattles) run() {
    for {
        select {
            case c := <- pbs.toAdd:
                // try to find a match
                matched := false
                for pb := range pbs.data {
                    if c.trainer == pb.trainer {
                        break
                    }
                    if c.CloseTo(pb) {
                        log.Println("matched")
                        delete(pbs.data, pb)
                        battle := Battle{conn1: pb.conn, conn2: c.conn}
                        c.trainer.battling = true
                        pb.trainer.battling = true
                        go battle.start(pb.Trainer_id, c.Trainer_id)
                        matched = true
                        break
                    }
                }
                if !matched {
                    pbs.data[c] = true
                }

                // remove old pending battles
                now := time.Now()
                for pb := range pbs.data {
                    if now.Sub(pb.createdTime).Seconds() > 60 {
                        delete(pbs.data, pb)
                        pb.remove()
                    }
                }

            case c := <- pbs.toRemove:
                delete(pbs.data, c)
                c.remove()
        }
    }
}

func (battle *Battle) start(trainer_id1 string, trainer_id2 string) {
    log.Println("Battle starting between", trainer_id1, "and", trainer_id2)

    trainer1 := makeTrainer(trainer_id1)
    trainer2 := makeTrainer(trainer_id2)

    train1ToMove := rand.Float32() < 0.5
    lastAttackMsg := LastAttackMessage{}

    roundNum := 0
    for {
        log.Println("Round", roundNum)
        roundNum += 1
        train1ToMove = !train1ToMove

        if battle.conn1.trainer.isWiped() && battle.conn2.trainer.isWiped() {
            log.Println("TIE")
            trainer1.outbox <- makeBattleResult(2).toBytes()
            trainer2.outbox <- makeBattleResult(2).toBytes()
            break
        }
        if battle.conn2.trainer.isWiped() {
            log.Println("Winner:", trainer1.name)
            trainer1.outbox <- makeBattleResult(0).toBytes()
            trainer2.outbox <- makeBattleResult(1).toBytes()
            break
        }
        if battle.conn1.trainer.isWiped() {
            log.Println("Winner:", trainer2.name)
            trainer1.outbox <- makeBattleResult(1).toBytes()
            trainer2.outbox <- makeBattleResult(0).toBytes()
            break
        }

        state1, state2 := battle.getStates(train1ToMove, lastAttackMsg)
        log.Println("Sending states")
        trainer1.outbox <- state1.toBytes()
        trainer2.outbox <- state2.toBytes()

        log.Println("Waiting for action", train1ToMove)
        if train1ToMove {
            lastAttackMsg = battle.process(<-trainer1.action)
        } else {
            lastAttackMsg = battle.process(<-trainer2.action)
        }
    }
    trainer1.battling = false
    trainer2.battling = false
    for _, pokemon := range trainer1.pokemon {
        pokemon.heal()
    }
    for _, pokemon := range trainer2.pokemon {
        pokemon.heal()
    }
}

func (battle *Battle) process(action *ActionMessage) LastAttackMessage {
    log.Println("Processing", action)

    result := LastAttackMessage{}
    trainer := Trainer{}
    other_trainer := Trainer{}

    if action.trainer == battle.conn1.trainer {
        trainer = *battle.conn1.trainer
        other_trainer = *battle.conn2.trainer
    } else {
        trainer = *battle.conn2.trainer
        other_trainer = *battle.conn1.trainer
    }

    if trainer.pokemon[0].state.health > 0 && action.Attack >= 0 && action.Attack < len(trainer.pokemon[0].moves) {
        if trainer.pokemon[0].state.pp[action.Attack] > 0 {
            result.Multiplier = trainer.pokemon[0].attack(other_trainer.pokemon[0], action.Attack)
            result.Pokemon = trainer.pokemon[0].name
            result.Move = trainer.pokemon[0].moves[action.Attack].Name
        }
    } else if action.Switch >= 0 && action.Switch < len(trainer.pokemon) {
        if trainer.pokemon[action.Switch].state.health > 0 {
            tmp := trainer.pokemon[action.Switch]
            trainer.pokemon[action.Switch] = trainer.pokemon[0]
            trainer.pokemon[0] = tmp
        }
    }
    return result
}

func (battle *Battle) getStates(train1ToMove bool, lastAttackMsg LastAttackMessage) (state1 StateMessage, state2 StateMessage) {
    trainer1 := *battle.conn1.trainer
    trainer2 := *battle.conn2.trainer

    state1 = makeStateMessage(train1ToMove, trainer1, trainer2, lastAttackMsg)
    state2 = makeStateMessage(!train1ToMove, trainer2, trainer1, lastAttackMsg)
    return
}
