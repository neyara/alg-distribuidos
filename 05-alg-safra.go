package main

import (
	"fmt"
	"math"
	"sync"
)

/*
* Struct que representa cada mensagem
* Sender: Processo que enviou o token
* Sum: Acumula a soma dos contadores de cada processo
* Last: Indica se eh o ultimo token enviado no alg de tarry
 */
type Token struct {
	Sender string
	Sum    int
	Last   bool
}

/*
* Struct que representa cada mensagem
* Sender: Processo que enviou a mensagem
* Dist: Contem a menor distancia encontrada
 */
type Message struct {
	Sender string
	Dist   float64
}

/*
* Struct que representa cada processo
* Value: Identifica o processo. Ex: P, Q,...
* Dist: Distância local
* Edges: Mapa representando as arestas deste nó aos vizinhos e os respectivos pesos
* Notify: Canal que retém as mensagens enviadas/recebidas de cada processo
* Pass: Canal que retém o token enviado/recebido de cada processo
* Done: Canal que sinaliza quando o algoritmo de terminação executou pela ultima vez no processo
 */
type Node struct {
	Value   string
	Dist    float64
	Edges   map[string]float64
	Counter int
	Notify  chan Message
	Pass    chan Token
	Done    chan struct{}
}

// Função que cria um novo processo
func newNode(value string) *Node {
	return &Node{
		Value:  value,
		Dist:   math.Inf(1), // definie distancia inicial como infinito
		Notify: make(chan Message, 5),
		Pass:   make(chan Token, 5),
		Done:   make(chan struct{}),
	}
}

// Função que liga v aos processos vizinhos, dado o vizinho e o peso da aresta
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

func termination(termin *sync.WaitGroup, currentNode *Node, beginner bool, active *sync.WaitGroup, token Token, neighs []*Node) {

	defer termin.Done()

	var father Node
	round := 1

	nmap := make(map[string]*Node)
	for _, neigh := range neighs {
		nmap[neigh.Value] = neigh
	}

	if beginner {
		// Processo iniciador
		fmt.Printf("== Iniciando terminação do processo [%s] ==\n", currentNode.Value)

		for {
			if token.Sum != 0 {
				fmt.Printf("================== Round %d ==================\n", round)
				token = Token{currentNode.Value, currentNode.Counter, false}
				neighs[0].Pass <- token // repropaga o token para o primeiro vizinho
				fmt.Printf("[%s] Token sent to %s (TOKEN SUM = %d)\n", currentNode.Value, neighs[0].Value, token.Sum)

				size := len(neighs)
				for i := 1; i < size; i++ {
					tk := <-currentNode.Pass // espera o token voltar para iniciador para passa-lo para os demais vizinhos
					fmt.Printf("[%s] Token received from %s\n", currentNode.Value, tk.Sender)
					tk.Sender = currentNode.Value
					neighs[i].Pass <- tk
					fmt.Printf("[%s] Token sent to %s\n", currentNode.Value, neighs[i].Value)
				}
				token = <-currentNode.Pass
				fmt.Printf("[%s] Token received from %s (counter = %d)\n", currentNode.Value, token.Sender, currentNode.Counter)
				fmt.Printf("================== Token sum is %d ================== \n", token.Sum)
			} else {
				// Qnd o token tem soma zero, ele eh enviado uma ultima vez para todos os nós sinalizando
				// que os mesmos não precisam esperar mais pelo token
				fmt.Printf("== Soma do token igual a 0 ==\n")
				fmt.Printf("============= Sinalizando fim de Tarry ============\n")

				token = Token{currentNode.Value, 0, true}
				neighs[0].Pass <- token // repropaga o token para o primeiro vizinho

				size := len(neighs)
				for i := 1; i < size; i++ {
					tk := <-currentNode.Pass // espera o token voltar para iniciador para passa-lo para os demais vizinhos
					fmt.Printf("[%s] Last token received from %s\n", currentNode.Value, tk.Sender)
					tk.Sender = currentNode.Value
					neighs[i].Pass <- tk
					fmt.Printf("[%s] Last token sent to %s\n", currentNode.Value, neighs[i].Value)
				}
				token = <-currentNode.Pass
				fmt.Printf("[%s] Last token received from %s\n", currentNode.Value, token.Sender)
				close(currentNode.Done) // fechar esse canal indica que o processo não está mais enviando/recebendo token
				break
			}
			round = round + 1
		}

		fmt.Println("== Fim do algoritmo de terminação! ==")

	} else {
		// Processo não iniciador

		fmt.Printf("== Iniciando terminação do processo [%s] ==\n", currentNode.Value)

		for {
			tk := <-currentNode.Pass

			if tk.Last == false {
				father.Value = ""
				fmt.Printf("[%s] Token received from %s (contador = %d)\n", currentNode.Value, tk.Sender, currentNode.Counter)
				tk.Sum = tk.Sum + currentNode.Counter // atualiza soma do token

				for _, neigh := range neighs {
					if father.Value == "" {
						father = *nmap[tk.Sender]
						//fmt.Printf("* %s é pai de %s\n", father.Value, currentNode.Value)
					}

					active.Wait()

					// Entrega o token para o vizinho se ele não for o pai
					if father.Value != neigh.Value {
						tk.Sender = currentNode.Value
						neigh.Pass <- tk // repropaga o token
						fmt.Printf("[%s] Token sent to %s\n", currentNode.Value, neigh.Value)
						tk = <-currentNode.Pass
						fmt.Printf("[%s] Token received from %s (contador = %d)\n", currentNode.Value, tk.Sender, currentNode.Counter)
					}

					//active.Wait()
				}
				// Token volta para o pai depois de ter passado enviado para todos os vizinhos
				tk.Sender = currentNode.Value
				father.Pass <- tk
				fmt.Printf("[%s] Token sent to father %s\n", currentNode.Value, father.Value)
			} else {
				// Qnd o token tem flag Last == true, o processo entende que não eh necessario continuar
				// esperando pelo token
				father.Value = ""
				fmt.Printf("[%s] Last token received from %s\n", currentNode.Value, tk.Sender)
				tk.Sum = tk.Sum + currentNode.Counter // atualiza soma do token

				for _, neigh := range neighs {
					if father.Value == "" {
						father = *nmap[tk.Sender]
					}

					active.Wait()

					// Entrega o token para o vizinho se ele não for o pai
					if father.Value != neigh.Value {
						tk.Sender = currentNode.Value
						neigh.Pass <- tk // repropaga o token
						fmt.Printf("[%s] Last token sent to %s\n", currentNode.Value, neigh.Value)
						tk = <-currentNode.Pass
						fmt.Printf("[%s] Last token received from %s\n", currentNode.Value, tk.Sender)
					}
				}
				// Token volta para o pai depois de ter passado enviado para todos os vizinhos
				tk.Sender = currentNode.Value
				father.Pass <- tk
				fmt.Printf("[%s] Last token sent to father %s\n", currentNode.Value, father.Value)
				close(currentNode.Done) // fechar esse canal indica que o processo não está mais enviando/recebendo token
				break
			}

		}

	}

}

func receiveRemainingMessages(beginnerNode *Node) {
	for {
		<-beginnerNode.Notify
		beginnerNode.Counter = beginnerNode.Counter - 1 //currentNode.Counter.dec()
		fmt.Printf("[%s] Receiving remaining messages (Counter = %d)\n", beginnerNode.Value, beginnerNode.Counter)
	}
}

func process(w *sync.WaitGroup, currentNode *Node, beginner bool, neighs ...*Node) {

	defer w.Done()

	var termin sync.WaitGroup // waitgroup que indica quando o alg de terminação finaliza
	var active sync.WaitGroup // waitgroup que quando o processo esta em estado passivo/ativo

	if beginner {
		// Processo iniciador

		currentNode.Dist = 0
		message := Message{Dist: currentNode.Dist, Sender: currentNode.Value}

		for _, neigh := range neighs {
			neigh.Notify <- message                       // envia msg para cada vizinho
			currentNode.Counter = currentNode.Counter + 1 // incrementa contador apos envio de msg
			fmt.Printf("[%s] Sending message to %s (Counter = %d)\n", currentNode.Value, neigh.Value, currentNode.Counter)
		}

		token := Token{currentNode.Value, currentNode.Counter, false} // inicializa o token
		termin.Add(1)
		go termination(&termin, currentNode, true, &active, token, neighs)

		// contando as mensagens recebidas
		go receiveRemainingMessages(currentNode)

	} else {
		// Processo não iniciador

		termin.Add(1)
		go termination(&termin, currentNode, false, &active, Token{}, neighs)

		for {
			done := false
			select {
			case msg := <-currentNode.Notify: // caso o processo receba alguma mensagem...

				active.Add(1) // processo esta no estado ACTIVE

				currentNode.Counter = currentNode.Counter - 1 // currentNode.Counter.dec()
				fmt.Printf("[%s] Receiving message from %s (Counter = %d)\n", currentNode.Value, msg.Sender, currentNode.Counter)

				newDist := msg.Dist + currentNode.Edges[msg.Sender]
				if newDist < currentNode.Dist {
					currentNode.Dist = newDist
					// fmt.Printf("[%s] New father = %s\n", currentNode.Value, currentNode.Father.Value)
					for _, neigh := range neighs {
						// notifica os vizinhos, exceto o pai
						if neigh.Value != msg.Sender {
							currentNode.Counter = currentNode.Counter + 1 // currentNode.Counter.inc()
							fmt.Printf("[%s] Sending message to %s (Counter = %d)\n", currentNode.Value, neigh.Value, currentNode.Counter)
							neigh.Notify <- Message{Dist: currentNode.Dist, Sender: currentNode.Value}
						}
					}
				}

				active.Done() // processo esta no estado PASSIVE

			case <-currentNode.Done: // caso o processo nao tenha mais mensagens para enviar/receber...
				done = true
				break
			}

			if done == true {
				//fmt.Printf("[%s] DONE\n", currentNode.Value)
				break
			}

		}

	}

	termin.Wait()
	fmt.Printf("[%s] PROCESS FINISHED - Chandy Finished!\n", currentNode.Value)

}

func main() {

	/*p := newNode("P")
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
	go process(&w, q, false, p, r) // Q

	w.Add(1)
	go process(&w, r, false, p, q, s, t) // R

	w.Add(1)
	go process(&w, t, false, r, s) // T

	w.Add(1)
	go process(&w, s, false, r, t) // S

	w.Add(1)
	go process(&w, p, true, q, r) // P

	w.Wait()*/

	p := newNode("P")
	q := newNode("Q")
	r := newNode("R")
	s := newNode("S")
	t := newNode("T")

	p.connect(q, 2) // já cobre o caso q.connect(p, 2)
	p.connect(r, 2)
	q.connect(r, 2)
	r.connect(t, 1)
	r.connect(s, 1)
	t.connect(s, 1)

	var w sync.WaitGroup // waitgroup que indica quando o alg de chandy finaliza

	w.Add(1)
	go process(&w, q, false, p, r) // Q

	w.Add(1)
	go process(&w, r, false, p, q, s, t) // R

	w.Add(1)
	go process(&w, t, false, r, s) // T

	w.Add(1)
	go process(&w, s, false, r, t) // S

	w.Add(1)
	go process(&w, p, true, q, r) // P

	w.Wait()
}
