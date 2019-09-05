// Implementação do algoritmo de travessia de Tarry
package main

import (
	"fmt"
	"sync"
)

type Token struct {
	Sender string
}

type Neighbour struct {
	Id   string
	Pass chan Token
}

func newNode(value string) *Neighbour {
	return &Neighbour{
		Id:   value,
		Pass: make(chan Token, 1),
	}
}

func redirect(in chan Token, neigh *Neighbour) {
	token := <-neigh.Pass
	in <- token
}

func process(w *sync.WaitGroup, currentNode *Neighbour, token Token, neighs ...*Neighbour) {
	var pai Neighbour

	defer w.Done()

	nmap := make(map[string]*Neighbour)
	for _, neigh := range neighs {
		nmap[neigh.Id] = neigh
	}

	if token.Sender == "init" {
		// Processo iniciador
		fmt.Printf("* %s é raiz.\n", currentNode.Id)
		token.Sender = currentNode.Id
		neighs[0].Pass <- token
		size := len(neighs)
		for i := 1; i < size; i++ {
			tk := <-currentNode.Pass
			fmt.Printf("From %s to %s\n", tk.Sender, currentNode.Id)
			tk.Sender = currentNode.Id
			neighs[i].Pass <- tk
		}
		tk := <-currentNode.Pass
		fmt.Printf("From %s to %s\n", tk.Sender, currentNode.Id)
		fmt.Println("Fim!")
	} else {
		// Processo não iniciador
		tk := <-currentNode.Pass
		fmt.Printf("From %s to %s\n", tk.Sender, currentNode.Id)
		for _, neigh := range neighs {
			if pai.Id == "" {
				pai = *nmap[tk.Sender]
				fmt.Printf("* %s é pai de %s\n", pai.Id, currentNode.Id)
			}
			// Entrega o token para o vizinho se ele não for o pai
			if pai.Id != neigh.Id {
				tk.Sender = currentNode.Id
				neigh.Pass <- tk
				tk = <-currentNode.Pass
				fmt.Printf("From %s to %s\n", tk.Sender, currentNode.Id)
			}
		}
		// Token volta para o pai depois de ter passado enviado para todos os vizinhos
		tk.Sender = currentNode.Id
		pai.Pass <- tk
	}

}

func main() {

	pNode := newNode("P")
	qNode := newNode("Q")
	rNode := newNode("R")
	sNode := newNode("S")
	wNode := newNode("W")

	var w sync.WaitGroup

	w.Add(1)
	go process(&w, wNode, Token{}, pNode, sNode)

	w.Add(1)
	go process(&w, sNode, Token{}, pNode, wNode)

	w.Add(1)
	go process(&w, rNode, Token{}, qNode, pNode)

	w.Add(1)
	go process(&w, qNode, Token{}, rNode)

	w.Add(1)
	go process(&w, pNode, Token{"init"}, wNode, sNode, rNode)

	w.Wait()
}
