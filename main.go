package main

import (
    "os"
    "github.com/pushtisonawala/chaos-fact-checker/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}