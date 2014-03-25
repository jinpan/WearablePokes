package main

import (
    "encoding/json"
)

type ActionMessage struct {
    trainer *Trainer
    Attack int `json:"attack"`
    Switch int `json:"switch"`
}

type BattleResult struct {
    Outcome string `json:"outcome"`
    DMoney int `json:"dMoney"`
}

type StateMessage struct {
    MyPokemon []PokemonMessage `json:"my_pokemon"`
    OtherPokemon PokemonMessage `json:"other_pokemon"`
    MyMove bool `json:"my_move"`
}

type PokemonMessage struct {
    Id string `json:"id"`
    Name string `json:"name"`
    Level uint `json:"level"`
    Health uint `json:"health"`
    Moves []MoveMessage `json:"moves"`
}

type MoveMessage struct {
    Name string `json:"name"`
    PP uint `json:"pp"`
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

func makeStateMessage(toMove bool, me Trainer, opponent Trainer) StateMessage {
    sm := StateMessage{MyMove: toMove, MyPokemon: make([]PokemonMessage, len(me.pokemon))}
    for idx, pokemon := range me.pokemon {
        sm.MyPokemon[idx] = makePokemonMessage(pokemon, true)
    }
    sm.OtherPokemon = makePokemonMessage(opponent.pokemon[0], false)
    return sm
}

func makePokemonMessage(pokemon Pokemon, full bool) PokemonMessage {
    msg := PokemonMessage{Id: pokemon.base.id, Name: pokemon.name,
                          Level: pokemon.level, Health: pokemon.state.health}
    if full {
        for idx, move := range pokemon.moves {
            msg.Moves = append(msg.Moves, MoveMessage{Name: move.Name, PP: pokemon.state.pp[idx]})
        }
    }

    return msg
}

