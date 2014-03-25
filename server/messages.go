package main

import (
    "encoding/json"
)

type ActionMessage struct {
    trainer *Trainer
    Attack int `json:"attack"`
    Switch int `json:"switch"`
}

type LastAttackMessage struct {
    Pokemon string `json:"pokemon"`
    Move string `json:"move"`
    Multiplier float64 `json:"multiplier"`
}

type BattleResult struct {
    Outcome string `json:"outcome"`
    DMoney int `json:"dMoney"`
}

type StateMessage struct {
    MyPokemon []PokemonMessage `json:"pokemon"`
    OtherPokemon PokemonMessage `json:"other_pokemon"`
    MyMove bool `json:"my_move"`
    LastAttack LastAttackMessage `json:"last_attack"`
}

type PokemonMessage struct {
    Id string `json:"id"`
    Name string `json:"name"`
    Level uint `json:"level"`
    Health uint `json:"hp"`
    MaxHealth uint `json:"maxhp"`
    Moves []MoveMessage `json:"moves"`
}

type MoveMessage struct {
    Name string `json:"name"`
    PP uint `json:"pp"`
    MaxPP uint `json:"maxpp"`
    Power uint `json:"power"`
    Type string `json:"type"`
}


func makeBattleResult(state uint) BattleResult {
    br := BattleResult{DMoney: 0}
    if state == 0 {
        br.Outcome = "Won"
    } else if state == 1 {
        br.Outcome = "Lost"
    } else {  // state == 2
        br.Outcome = "Tie"
    }
    return br
}

func (br BattleResult) toBytes() []byte {
    msg, _ := json.Marshal(br)
    return msg
}

func (sm StateMessage) toBytes() []byte {
    msg, _ := json.Marshal(sm)
    return msg
}

func makeStateMessage(toMove bool, me Trainer, opponent Trainer, lastAttackMsg LastAttackMessage) StateMessage {
    sm := StateMessage{MyMove: toMove, MyPokemon: make([]PokemonMessage, len(me.pokemon))}
    for idx, pokemon := range me.pokemon {
        sm.MyPokemon[idx] = makePokemonMessage(pokemon, true)
    }
    sm.OtherPokemon = makePokemonMessage(opponent.pokemon[0], false)
    sm.LastAttack = lastAttackMsg
    return sm
}

func makePokemonMessage(pokemon Pokemon, full bool) PokemonMessage {
    msg := PokemonMessage{Id: pokemon.base.id, Name: pokemon.name,
                          Level: pokemon.level, Health: pokemon.state.health,
                          MaxHealth: pokemon.maxHealth}
    if full {
        for idx, move := range pokemon.moves {
            msg.Moves = append(msg.Moves, MoveMessage{Name: move.Name,
                                                      PP: pokemon.state.pp[idx],
                                                      MaxPP: pokemon.moves[idx].MaxPP,
                                                      Power: pokemon.moves[idx].Power,
                                                      Type: pokemon.moves[idx].TypeString})
        }
    }

    return msg
}

