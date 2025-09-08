package commands_test

import (
    "testing"

    cmds "go-redis/internal/usecases/redis/commands"
)

func TestParseRespToKeyValue_BasicSet(t *testing.T) {
    uc := cmds.NewParseRespToKeyValueUseCase()
    resp := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

    m, err := uc.Handle(resp)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got := m["SET"]; got != "foo" {
        t.Fatalf("unexpected parsed value: %v", got)
    }
}

