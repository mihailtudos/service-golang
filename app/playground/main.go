package main

import (
	"fmt"
	"github.com/mihailtudos/service3/foundation/keystore"
	"log"
	"os"
)

func main() {

	ks, err := keystore.NewFS(os.DirFS("zarf/keys/"))
	if err != nil {
		panic(err)
	}

	fmt.Println(ks)

	pk, err := ks.PublicKey("456F21BD-1296-449A-9C2E-85A92092E966")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pk)
}
