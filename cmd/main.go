package main

import (
    "encoding/json"
    "os"
    "log"
    "fmt"
    "flag"

    "example.com/jsonp"
)

var mode = flag.String("format", "json", "output format [json/raw]")

func main() {
    flag.Parse()

    lexer := jsonp.NewLexer(os.Stdin)

    tokens, err := lexer.Lex()

    if err != nil {
        log.Fatal(err)
    }

    if *mode == "json" {
        value, err := json.Marshal(tokens)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf(string(value))
        return
    }

    for _, token := range tokens {
        fmt.Printf("%#v\n", token)
    }
}
