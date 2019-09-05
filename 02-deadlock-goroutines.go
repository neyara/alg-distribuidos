package main

import (
	"fmt"
	"sync"
)

/*
* Struct que representa cada processo
* Value: Identifica o processo. Ex: P, Q,...
* From: Representa o processo que levou ate o processo atual
* VisitedTime: Representa o tempo quando o processo é visitado pela primeira vez
* FinishedTime: Representa o tempo quando o processo é visitado pela ultima vez
* Authorized: Canal que indica quando o processo pode visitar o(s) proximo(s) filho(s)
* Done: Canal que indica que o(s) filho(s) terminou a execução do algoritmo
 */
type Node struct {
	Value        string
	From         *Node
	VisitedTime  int
	FinishedTime int
	Authorized   chan bool
	Done         chan bool
}

var count int
var dList [][]string

// cria um novo processo
func newNode(value string) *Node {
	return &Node{
		Value:      value,
		Authorized: make(chan bool),
		Done:       make(chan bool),
	}
}

// incrementa os tempos dos processos
func incrementTime(node *Node) {
	count = count + 1
	if node.VisitedTime == 0 {
		node.VisitedTime = count
	} else {
		node.FinishedTime = count
	}
}

// armazena o(s) caminho(s) onde teve ocorrencia de deadlock
func getDeadlockPath(neigh, currentNode *Node) {
	dSlice := make([]string, 0)
	// adiciona os nós que se encontram nos extremos da aresta de retorno
	dSlice = append(dSlice, neigh.Value, currentNode.Value)
	nextNode := currentNode.From
	// recupera os nós intermediários que formam o deadlock (ciclo)
	for nextNode.Value != neigh.Value {
		dSlice = append(dSlice, nextNode.Value)
		nextNode = nextNode.From
	}
	dList = append(dList, dSlice)
}

func process(w *sync.WaitGroup, currentNode *Node, beginner bool, neighs ...*Node) {

	defer w.Done()

	for _, neigh := range neighs {
		if neigh.From == nil {
			neigh.From = currentNode
		}
	}

	if beginner {
		// Processo iniciador
		fmt.Printf("* %s é raiz.\n", currentNode.Value)
		size := len(neighs)
		incrementTime(currentNode) // Incrementa o VisitedTime

		fmt.Printf("(Enviando) %s[%d/%d] -> %s[%d/%d]\n", currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime, neighs[0].Value, neighs[0].VisitedTime, neighs[0].FinishedTime)
		neighs[0].Authorized <- true

		for i := 1; i < size; i++ {
			<-currentNode.Done
			fmt.Printf("(Enviando) %s[%d/%d] -> %s[%d/%d]\n", currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime, neighs[i].Value, neighs[i].VisitedTime, neighs[i].FinishedTime)
			neighs[i].Authorized <- true
		}

		<-currentNode.Done         // Espera o último filho acabar a execução
		incrementTime(currentNode) // Incrementa o FinishedTime
		fmt.Printf("Processo (%s[%d/%d]) finalizado.\n", currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime)
		fmt.Println("Fim!")
	} else {
		// Processo não iniciador
		<-currentNode.Authorized
		incrementTime(currentNode) // Incrementa o VisitedTime
		fmt.Printf("(Recebendo) %s[%d/%d] -> %s[%d/%d]\n", currentNode.From.Value, currentNode.From.VisitedTime, currentNode.From.FinishedTime, currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime)

		for _, neigh := range neighs {

			if neigh.VisitedTime != 0 {
				fmt.Printf("%s já foi visitado!\n", neigh.Value)
				if neigh.FinishedTime == 0 {
					fmt.Printf("# DEADLOCK - Aresta de retorno entre os nós %s e %s\n", currentNode.Value, neigh.Value)
					getDeadlockPath(neigh, currentNode)
				}
			} else {
				fmt.Printf("(Enviando) %s[%d/%d] -> %s[%d/%d]\n", currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime, neigh.Value, neigh.VisitedTime, neigh.FinishedTime)
				neigh.Authorized <- true
				<-currentNode.Done
			}

		}

		incrementTime(currentNode) // Incrementa o FinishedTime
		fmt.Printf("Processo (%s[%d/%d]) finalizado. Voltando para o pai (%s[%d/%d])...\n", currentNode.Value, currentNode.VisitedTime, currentNode.FinishedTime, currentNode.From.Value, currentNode.From.VisitedTime, currentNode.From.FinishedTime)
		currentNode.From.Done <- true // Avisa ao pai que as visitas foram finalizadas
	}

}

func main() {

	dList = make([][]string, 0)

	/*p := newNode("P")
	q := newNode("Q")
	r := newNode("R")
	n := newNode("N")
	s := newNode("S")
	t := newNode("T")

	var w sync.WaitGroup

	w.Add(1)
	go process(&w, t, false, n) // T

	w.Add(1)
	go process(&w, q, false, r) // Q

	w.Add(1)
	go process(&w, r, false) // R

	w.Add(1)
	go process(&w, n, false, s) // N

	w.Add(1)
	go process(&w, s, false, t) // S

	w.Add(1)
	go process(&w, p, true, q, n) // P

	w.Wait()*/

	p := newNode("P")
	q := newNode("Q")
	r := newNode("R")
	s := newNode("S")

	var w sync.WaitGroup

	w.Add(1)
	go process(&w, s, false) // S

	w.Add(1)
	go process(&w, q, false, s) // Q

	w.Add(1)
	go process(&w, r, false, s) // R

	w.Add(1)
	go process(&w, p, true, q, r) // P

	w.Wait()

	// Impressao dos deadlocks detectados, se existirem
	if len(dList) > 0 {
		fmt.Println("Lista de deadlocks encontrados:")
		for i, element := range dList {
			fmt.Printf("(%d) %v\n", i+1, element)
		}
	} else {
		fmt.Println("O sistema não possui deadlocks.")
	}

}
