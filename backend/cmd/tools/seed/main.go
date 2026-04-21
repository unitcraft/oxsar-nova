// Command seed — utility, сеящая в БД первого игрока и стартовую планету.
//
// TODO: полноценная CLI с подкомандами (reseed, resync-artefacts,
// remove-inactive и т.д.). Сейчас — минимальный helper для ручного
// dev-старта до UI-регистрации.
package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "seed failed:", err)
		os.Exit(1)
	}
}

func run(_ context.Context) error {
	fmt.Println("seed: not implemented yet — use POST /api/auth/register")
	return nil
}
