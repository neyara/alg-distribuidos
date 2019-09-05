package main

import (
	"fmt"
	"sync"
)

/*
* Struct que representa a informacao passada por cada processo
* Sender: Identifica o processo que enviou a informacao
* Value: Valor do processo eleito ate o momento
* ChosenNodeId: Identifica o processo eleito ate o momento
 */
type Info struct {
	Sender          string
	ChosenNodeId    string
	ChosenNodeValue int
}

/*
* Struct que representa cada processo
* Id: Identifica o processo. Ex: P, Q,...
* Value: Valor do processo atual
* Notify: Canal que retém a informação passada por cada processo
* AuthorizedBy: Canal que indica quando o processo pode visitar o(s) proximo(s) filho(s)
 */
type Node struct {
	Id           string
	Value        int
	Notify       chan Info
	AuthorizedBy chan string
}

// Função que cria um novo processo com seu respectivo valor
func newNode(id string, value int) *Node {
	return &Node{
		Id:           id,
		Value:        value,
		Notify:       make(chan Info, 1),
		AuthorizedBy: make(chan string, 1),
	}
}

func process(w *sync.WaitGroup, currentNode *Node, beginner bool, neighs ...*Node) {

	defer w.Done()

	var chosenNode Node
	chosenNode.Id = currentNode.Id
	chosenNode.Value = currentNode.Value

	nmap := make(map[string]*Node)
	for _, neigh := range neighs {
		nmap[neigh.Id] = neigh
	}

	if beginner {
		// Processo iniciador
		fmt.Printf("* %s é o processo iniciador.\n", currentNode.Id)

		// Notifica todos aos nós vizinhos que eles podem visitar seus filhos
		for _, neigh := range neighs {
			fmt.Printf("[%s] Authorizing process %s to continue\n", currentNode.Id, neigh.Id)
			neigh.AuthorizedBy <- currentNode.Id
		}

		// Recebe o nó eleito por cada filho
		for i := 0; i < len(neighs); i++ {
			infoReceived := <-currentNode.Notify
			if infoReceived.ChosenNodeValue > chosenNode.Value {
				fmt.Printf("[%s] A new chosen process was found by %s\n", currentNode.Id, infoReceived.Sender)
				chosenNode.Id = infoReceived.ChosenNodeId
				chosenNode.Value = infoReceived.ChosenNodeValue
			}
		}

		// O processo é finalizado e o novo processo iniciador do sistema é eleito
		fmt.Printf("The new beginner will be: %s (%d)\n", chosenNode.Id, chosenNode.Value)

	} else {
		// Processo não iniciador

		numChildren := 0 // variavel que armazena o numero de filhos do nó
		sender := <-currentNode.AuthorizedBy
		fmt.Printf("Iniciando processo %s...\n", currentNode.Id)

		// Notifica todos aos nós vizinhos que eles podem visitar seus filhos
		for _, neigh := range neighs {
			if neigh.Id != sender {
				fmt.Printf("[%s] Authorizing process %s to continue\n", currentNode.Id, neigh.Id)
				neigh.AuthorizedBy <- currentNode.Id
				numChildren = numChildren + 1
			}
		}

		// Se o nó não possui filhos ele se elege como eleito temporario
		if numChildren == 0 {

			chosenNode.Id = currentNode.Id
			chosenNode.Value = currentNode.Value

		} else {

			// Recebe o nó eleito por cada filho
			for i := 0; i < numChildren; i++ {
				infoReceived := <-currentNode.Notify
				if infoReceived.ChosenNodeValue > chosenNode.Value {
					fmt.Printf("[%s] A new chosen process was found by %s\n", currentNode.Id, infoReceived.Sender)
					chosenNode.Id = infoReceived.ChosenNodeId
					chosenNode.Value = infoReceived.ChosenNodeValue
				}
			}

		}

		// Informa ao pai qual o processo eleito naquela subarvore
		father := *nmap[sender]
		father.Notify <- Info{currentNode.Id, chosenNode.Id, chosenNode.Value}

		// O processo atual é finalizado
		fmt.Printf("[%s] Notifying the father (%s) chosen process is %s(%d)\n", currentNode.Id, sender, chosenNode.Id, chosenNode.Value)
	}

}

func main() {

	p := newNode("P", 3)
	q := newNode("Q", 11)
	r := newNode("R", 8)
	s := newNode("S", 7)
	t := newNode("T", 20)
	v := newNode("V", 10)

	var w sync.WaitGroup

	w.Add(1)
	go process(&w, t, false, q) // T

	w.Add(1)
	go process(&w, q, false, p, t, s) // Q

	w.Add(1)
	go process(&w, r, false, p, v) // R

	w.Add(1)
	go process(&w, s, false, q) // S

	w.Add(1)
	go process(&w, v, false, r) // V

	w.Add(1)
	go process(&w, p, true, q, r) // P

	w.Wait()
}
