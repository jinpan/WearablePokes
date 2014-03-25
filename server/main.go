package main

import (
    "encoding/json"
    "flag"
    "log"
    "net/http"
    "os"
)

var (
    pendingBattles = PendingBattles{data: make(map[*PendingBattle] bool),
                                    toAdd: make(chan *PendingBattle),
                                    toRemove: make(chan *PendingBattle)}
    addr = flag.String("addr", ":10914", "http service address")
    types_fp = flag.String("types_fp", "json/types.json", "path to types data")
    moves_fp = flag.String("moves_fp", "json/moves.json", "path to moves data")
    pokemon_fp = flag.String("pokemon_fp", "json/pokemon.json", "path to pokemon data")

    raw_types = make(map[string] map[string] float64)
    types = make(map[string] Type)
    moves = make(map[string] *PokemonMove)
    base_pokemon = make(map[string] *BasePokemon)

    trainers = make(map[string] *Trainer)

    jin = Trainer{}
    edwin = Trainer{}
)

func parseFiles() {
    types_file, _ := os.OpenFile(*types_fp, os.O_RDONLY, 0644)
    json.NewDecoder(types_file).Decode(&raw_types)

    for key := range raw_types {
        types[key] = Type{name: key,
                          strongVS: make(map[*Type] bool),
                          weakVS: make(map[*Type] bool),
                          ineffectiveVS: make(map[*Type] bool)}
    }
    for key, val := range raw_types {
        myType := types[key]
        for name, score := range val {
            otherType := types[name]
            if score == 2.0 {
                myType.strongVS[&otherType] = true
            } else if score == 0.5 {
                myType.weakVS[&otherType] = true
            } else if score == 0.0 {
                myType.ineffectiveVS[&otherType] = true
            }
        }
    }

    moves_file, _ := os.OpenFile(*moves_fp, os.O_RDONLY, 0644)
    json.NewDecoder(moves_file).Decode(&moves)
    for _, move := range moves {
        moveType := types[move.TypeString]
        move.moveType = &moveType
    }

    pokemon_file, _ := os.OpenFile(*pokemon_fp, os.O_RDONLY, 0664)
    json.NewDecoder(pokemon_file).Decode(&base_pokemon)
    for id, base := range base_pokemon {
        base.id = id
        pokemonType := types[base.TypeString]
        base.pokemonType = &pokemonType
    }
}

func main() {
    parseFiles()
    jin = Trainer{name: "Jin", id: "1", pokemon: []Pokemon{
        makePokemon("151", "Mew", 100),
        makePokemon("149", "Dragonite", 100),
        makePokemon("145", "Zapdos", 100),
        makePokemon("3", "Venusaur", 100),
        makePokemon("6", "Blastoise", 100),
        makePokemon("9", "Charizard", 100)},
        action: make(chan *ActionMessage),
        outbox: make(chan []byte),
        connections: make(map[*BattleConnection] bool),
        battling: false,
    }
    edwin = Trainer{name: "Edwin", id: "2", pokemon: []Pokemon{
        makePokemon("4", "Charmander", 100),
        makePokemon("42", "COOL GUY", 100),
        makePokemon("88", "Sweet Bro", 100),
        makePokemon("12", "Chocolate", 100),
        makePokemon("55", "Google Glass", 100),
        makePokemon("56", "Myo", 100),
        makePokemon("150", "MISSINGNO", 100)},
        action: make(chan *ActionMessage),
        outbox: make(chan []byte),
        connections: make(map[*BattleConnection] bool),
        battling: false,
    }

    trainers["1"] = &jin
    trainers["2"] = &edwin

    go jin.run()
    go edwin.run()

    go pendingBattles.run()
    http.HandleFunc("/battle", battleHandler)

    if err := http.ListenAndServe(*addr, nil); err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

