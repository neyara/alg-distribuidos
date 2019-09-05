package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

/*
* Struct que representa cada processo
* Value: Identifica o processo. Ex: P, Q,...
* Dist: Distência local
* Father: Pai do processo
* Edges: Mapa representando as arestas deste nó aos vizinhos e os respectivos pesos
* Notify: Canal que retém as mensagens enviadas/recebidas de cada processo
 */
type Node struct {
	Value  string
	Dist   float64
	Father *Node
	Edges  map[string]float64
	Notify chan *Node
}

/*
* Struct para armazenar o resultado final de cada processo
* Path: caminho até o processo iniciador
* Dist: menor distância
 */
type Result struct {
	Path string
	Dist float64
}

// Lista que retém os caminhos resultantes e as menores distências
// O resultado é apresentado apenas quando o algoritmo finaliza
var dList []Result

// Função que cria um novo processo
func newNode(value string) *Node {
	return &Node{
		Value:  value,
		Dist:   math.Inf(1), // definie dist inicial como infinito
		Notify: make(chan *Node, 2),
	}
}

// Função que liga processos vizinhos, dado o vizinho e o peso da aresta
func (v *Node) connect(neigh *Node, weight float64) {
	if v.Edges == nil {
		v.Edges = make(map[string]float64)
	}
	if neigh.Edges == nil {
		neigh.Edges = make(map[string]float64)
	}
	v.Edges[neigh.Value] = weight
	neigh.Edges[v.Value] = weight
}

// Função que armazena o caminho do precesso "node" até o iniciador na lista dList
func addPathToList(node *Node) {
	distToBeginner := node.Dist
	path := node.Value
	for node.Father != nil {
		path += " -> " + node.Father.Value
		node = node.Father
	}
	dList = append(dList, Result{Path: path, Dist: distToBeginner})
}

func process(w *sync.WaitGroup, currentNode *Node, beginner bool, neighs ...*Node) {

	defer w.Done()

	if beginner {
		// Processo iniciador
		fmt.Printf("* %s é o processo iniciador.\n", currentNode.Value)

		currentNode.Dist = 0

		for _, neigh := range neighs {
			fmt.Printf("(Sending) [%s] -> %s\n", currentNode.Value, neigh.Value)
			neigh.Notify <- currentNode // notifica cada vizinho
		}

	} else {
		// Processo não iniciador

		fmt.Printf("Iniciando processo %s...\n", currentNode.Value)

		started := false // indica quando o processo recebeu uma msg pela 1ª vez
		attempts := 0    // indica o numero de tentativas do processo ao receber uma msg
		for {
			select {
			case sender := <-currentNode.Notify:
				started = true // quando o processo é notificado pela primeira vez, indicamos que as tentativas podem começar a ser contabilizadas
				attempts = 0   // o nº de tentativas é zerado quando o processo recebe uma mensagem de algum outro processo
				fmt.Printf("(Receiving) %s -> [%s]\n", sender.Value, currentNode.Value)
				newDist := sender.Dist + currentNode.Edges[sender.Value]
				if newDist < currentNode.Dist {
					currentNode.Dist = newDist
					currentNode.Father = sender
					for _, neigh := range neighs {
						// notifica os vizinhos, exceto o pai
						if neigh.Value != sender.Value {
							fmt.Printf("(Sending) [%s] -> %s\n", currentNode.Value, neigh.Value)
							neigh.Notify <- currentNode
						}
					}
				}
			default:
				// O processo deve receber pelo menos uma notificacao para comecar a contagem das tentativas
				// por isso, o uso da variavel "started"
				if started {
					attempts = attempts + 1
					//fmt.Printf("[%s] Falha ao receber dados. Tentativa #%d...\n", currentNode.Value, attempts)
					time.Sleep(1 * time.Second)
				}
			}

			// quando o numero de tentativas eh 2, o processo deixa de escutar o canal de mensagens
			// OBS: esse metodo so funciona para exemplos especificos e foi usado apenas para
			//      forçar a terminação do algoritmo e confirmar a saida esperada
			if attempts == 2 {
				addPathToList(currentNode)
				break
			}

		}

		// Algoritmo original do chandy-misra
		/*for {
			sender := <-currentNode.Notify
			fmt.Printf("(Receiving) %s -> [%s]\n", sender.Value, currentNode.Value)
			newDist := sender.Dist + currentNode.Edges[sender.Value]
			if newDist < currentNode.Dist {
				currentNode.Dist = newDist
				currentNode.Father = sender
				for _, neigh := range neighs {
					// notifica os vizinhos, exceto o pai
					if neigh.Value != sender.Value {
						neigh.Notify <- currentNode
					}
				}
			}
		}*/

		// Após o número máximo de tentativas, o processo é finalizado
		fmt.Printf("Processo (%s) finalizou o envio de mensagens...\n", currentNode.Value)
	}

}

func main() {

	dList = make([]Result, 0) // Lista para armazenar os caminhos resultantes e as menores distâncias

	p := newNode("P")
	q := newNode("Q")
	r := newNode("R")
	s := newNode("S")
	t := newNode("T")

	p.connect(q, 1) // já cobre o caso q.connect(p, 2)
	p.connect(r, 3)
	q.connect(r, 1)
	r.connect(t, 4)
	r.connect(s, 1)
	t.connect(s, 1)

	var w sync.WaitGroup

	w.Add(1)
	go process(&w, t, false, r, s) // T

	w.Add(1)
	go process(&w, q, false, p, r) // Q

	w.Add(1)
	go process(&w, r, false, p, q, s, t) // R

	w.Add(1)
	go process(&w, s, false, r, t) // S

	w.Add(1)
	go process(&w, p, true, q, r) // P

	w.Wait()

	// Impressao dos caminhos até o iniciador para cada processo
	// bem como da menor distância encontrada
	fmt.Println("Caminhos encontrados:")
	for _, element := range dList {
		fmt.Printf("%v\n", element)
	}
}
