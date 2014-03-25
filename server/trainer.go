package main

type Trainer struct {
    name string
    id string
    pokemon []Pokemon
    money uint64
}

func makeTrainer(id string) Trainer {
    if id == "1" {
        return jin
    } else if id == "2" {
        return edwin
    } else {
        return edwin
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

