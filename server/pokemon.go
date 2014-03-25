package main

import (
    "log"
)

type BasePokemon struct {
    id string
    Name string `json:"name"`

    Attack uint `json:"attack"`
    Defense uint `json:"defense"`

    TypeString string `json:"type"`
    pokemonType *Type `json:"type"`

    Moves []string `json:"moves"`
}

type Pokemon struct {
    name string
    base BasePokemon

    level uint
    maxHealth uint
    moves []PokemonMove

    state *PokemonState
}

type PokemonState struct {
    health uint
    pp []uint
}

type PokemonMove struct {
    Name string `json:"name"`
    MaxPP uint `json:"pp"`
    Power uint `json:"power"`
    TypeString string `json:"type"`
    moveType *Type
}

type Type struct {
    name string
    strongVS map[*Type] bool
    weakVS map[*Type] bool
    ineffectiveVS map[*Type] bool
}

func makePokemon(id string, name string, level uint) Pokemon {
    pokemon := Pokemon{name: name, level: level}
    base := base_pokemon[id]

    pokemon.maxHealth = uint(30.0 + float64(base.Defense) * (1 + float64(level) / 50.0))
    pokemon.base = *base

    pokemon.moves = make([]PokemonMove, len(base.Moves))
    pp := make([]uint, len(base.Moves))
    for idx, moveName := range base.Moves {
        move := moves[moveName]
        pokemon.moves[idx] = *move
        pp[idx] = moves[moveName].MaxPP
    }

    state := PokemonState{health: pokemon.maxHealth, pp: pp}
    pokemon.state = &state

    return pokemon
}

func (p Pokemon) heal() {
    p.state.health = p.maxHealth
    for idx := range p.moves {
        p.state.pp[idx] = p.moves[idx].MaxPP
    }
}

func (p1 Pokemon) attack(p2 Pokemon, move_idx int) (X float64, dmg uint) {
    move := p1.moves[move_idx]

    A := float64(p1.level)
    B := float64(p1.base.Attack)
    C := float64(move.Power)
    D := float64(p2.base.Defense)
    X = calcEffectiveness(move.moveType, p2.base.pokemonType)
    Y := 1.0
    if p1.base.pokemonType.name == move.moveType.name {
        Y = 1.5
    }

    dmg = uint(calcDmg(A, B, C, D, X, Y))
    log.Println("Damage:", dmg)
    if dmg > p2.state.health {
        p2.state.health = 0
    } else {
        p2.state.health -= dmg
    }
    p1.state.pp[move_idx] -= 1

    return
}


func calcEffectiveness(moveType *Type, pokemonType *Type) float64 {
    if moveType.strongVS[pokemonType] {
        return 2.0
    } else if moveType.weakVS[pokemonType] {
        return 0.5
    } else if moveType.ineffectiveVS[pokemonType] {
        return 0.0
    } else {
        return 1.0
    }
}

func calcDmg(A float64, B float64, C float64, D float64, X float64, Y float64) float64 {
    return (2 * A / 5 + 2) * B * C * X * Y / D / 30
}
