package main

import (
	"fmt"
	"strconv"
)

// Cluster configuration informamtion
const NodeNum = 3 // # of nodes
const PodNum = 6  // # of replicas
const alpha = 1.0 // ratio between processing delay and latency

type Node struct {	// Resource data type for the nodes
	Name string
	CPU  float64
	RAM  float64
	B    float64
}

type NodeDelay struct {	//	Records the corresponding latency to the node
	Name  string
	Delay float64
}

type NumOfCon struct {	// Records the number of containers(Pods) on the node
	Name   string
	Number float64
}

type Pod struct {	// Resource data type of the container(Pod)
	Name string
	CPU  float64
	RAM  float64
	B    float64
}

type FinalScore struct {	// Scoring result for the node-pod pairing
	NodeName string
	PodName  string
	Score    float64
}

type Binding struct{ PodName, NodeName string }

// Side Functions ------------------------------------------------------------------------------------------------------

func Sub(n Node, p Pod) Node { // Subtract the resource required from Pod
	if (n.CPU-p.CPU) < 0 || (n.RAM-p.RAM) < 400 || (n.B-p.B) < 0 { // should left enough RAM in case of OOM error
		return Node{n.Name, 0, 0, 0} // Set all resource to 0 to filter out this node
	} else {
		return Node{n.Name, n.CPU - p.CPU, n.RAM - p.RAM, n.B - p.B}
	}
}

func NormCPU(n [NodeNum]Node, c [NodeNum]NumOfCon, p Pod, s string) [NodeNum]Node { // Map the remaining CPU to processing overhead

	//---------------------------- S T R E S S - N G ----------------------------

	for i := 0; i < NodeNum; i++ { // Mapping for Stress-ng 30 sec
		fmt.Print(c[i], " ")

		if s == "current" {
			n[i].CPU = (n[i].CPU / (c[i].Number + 1)) + p.CPU // additional resources (divided by # of pods on the same node) + request resources
			fmt.Println("CPU share on", n[i].Name, ": ", n[i].CPU)
		} else if s == "last" && c[i].Number == 0 {
			n[i].CPU = ((n[i].CPU) / 1) + p.CPU // additional resources (divided by # of pods on the same node) + request resources
			fmt.Println("CPU share on", n[i].Name, ": ", n[i].CPU)
		} else if s == "last" && c[i].Number > 0 {
			n[i].CPU = ((n[i].CPU + p.CPU) / c[i].Number) + p.CPU // additional resources (divided by # of pods on the same node) + request resources
			fmt.Println("CPU share on", n[i].Name, ": ", n[i].CPU)
		}

		if n[i].CPU > 0 && n[i].CPU <= 0.25 { // Unit: Core
			n[i].CPU = 714 // Unit: ms
		} else if n[i].CPU > 0.25 && n[i].CPU <= 0.5 {
			n[i].CPU = 303
		} else if n[i].CPU > 0.5 && n[i].CPU <= 0.75 {
			n[i].CPU = 186
		} else if n[i].CPU > 0.75 && n[i].CPU <= 1 {
			n[i].CPU = 165
		} else if n[i].CPU > 1 && n[i].CPU <= 1.25 {
			n[i].CPU = 137
		} else if n[i].CPU > 1.25 && n[i].CPU <= 1.5 {
			n[i].CPU = 120
		} else { // Near full (>1.5)
			n[i].CPU = 110
		}
	}

	// //---------------------------- S Y S B E N C H ----------------------------
	// for i := 0; i < NodeNum; i++ { // Mapping for Sysbench 20 prime number
	// 	if n[i].CPU > 0 && n[i].CPU <= 0.1 { // Unit: Core
	// 		n[i].CPU = 95 // Unit: ms
	// 	} else if n[i].CPU > 0.1 && n[i].CPU <= 0.5 {
	// 		n[i].CPU = 42
	// 	} else if n[i].CPU > 0.5 && n[i].CPU <= 1 {
	// 		n[i].CPU = 21
	// 	} else if n[i].CPU > 1 && n[i].CPU <= 1.5 {
	// 		n[i].CPU = 3
	// 	} else { // Near full (=2)
	// 		n[i].CPU = 1
	// 	}
	// }

	fmt.Println()
	return n
}

func Min(a [NodeNum]float64) float64 { // get min value in an array
	var min = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
	}
	return min
}

func Max(a [NodeNum]float64) float64 { // get max value in an array
	var max = a[0]
	for _, value := range a {
		if value > max {
			max = value
		}
	}
	return max
}

// Main Function ------------------------------------------------------------------------------------------------------

func main() {
	var N [NodeNum]Node    // List of worker nodes
	var NSub [NodeNum]Node // List of worker nodes after subtraction
	var NLastSession [NodeNum]Node
	var D [NodeNum]NodeDelay // List of workers' average delay
	var P [PodNum]Pod        // List of Pods' required resources

	var C [NodeNum]NumOfCon

	var FS [len(N)]FinalScore

	var bind [PodNum]Binding // Store binding information Node/Pod

	var Fail [NodeNum]bool

	NSub = N // copy Node resource info to NSub for further operation (subtraction)

	//
	//
	//---------------------------------- I N I T I A L    V A L U E ----------------------------------( S T A R T )

	fmt.Println()
	fmt.Println("----------- Node Resource Status -----------")

	// Initial status for available resources on the nodes (NAME, CPU, RAM, Bandwidth)
	N[0] = Node{"worker-" + strconv.Itoa(1), 1.74, 1500, 30} // Itoa: int to string
	N[1] = Node{"worker-" + strconv.Itoa(2), 0.8, 1490, 30}
	N[2] = Node{"worker-" + strconv.Itoa(3), 0.9, 1210, 30}
	fmt.Println(N)
	fmt.Println()

	fmt.Println("----------- Node Delay Status -----------")

	// Initial status for "average delay" on the nodes
	D[0] = NodeDelay{"worker-" + strconv.Itoa(1), 220.0} // Itoa: int to string
	D[1] = NodeDelay{"worker-" + strconv.Itoa(2), 126.7}
	D[2] = NodeDelay{"worker-" + strconv.Itoa(3), 120.0}
	fmt.Println(D)
	fmt.Println()

	fmt.Println("----------- Pod Resource Reqest -----------")
	for j := 0; j < len(P); j++ {
		P[j] = Pod{"pod-" + strconv.Itoa(j+1), 0.25, 100, 5} // Itoa: int to string
		fmt.Println(P[j])
	}
	fmt.Println()

	for i := 0; i < len(N); i++ {
		C[i].Name = N[i].Name
		C[i].Number = 0
	}

	//---------------------------------- I N I T I A L    V A L U E  ----------------------------------( E N D )
	//
	//

	for j := 0; j < len(P); j++ {
		for i := 0; i < len(N); i++ { // Subtract the required resource from pod + Filtering
			NSub[i] = Sub(N[i], P[j]) // Pod index
		}

		count := 0
		for i := 0; i < len(N); i++ { // Count/Flag all unavailable nodes
			if NSub[i].RAM <= 0 { // RAM should have at least 400 (4GB*10%), 0 stands for node not suitable
				Fail[i] = true // Flag the Node index without enough resource to run the Pod (Filtering)
				count++
			} else {
				Fail[i] = false // Flag the Node index with enough resource to run the Pod
			}
		}

		if count == len(N) { // No available node in the cluster
			bind[j].NodeName = "Failed" // Pod index
			bind[j].PodName = P[j].Name // Pod index
			fmt.Println("No node is available for:", P[j].Name, "\n\n\n\n")

		} else { // At least one node is available
			//
			//
			//---------------------------------- S C O R I N G ----------------------------------
			NLastSession = NSub
			NSub = NormCPU(NSub, C, P[j], "current") // Normalize the node CPU to actual processing overhead(delay)
			NLastSession = NormCPU(NLastSession, C, P[j], "last")
			fmt.Println("----------- Node available CPU to processing overhead -----------")
			fmt.Println("Current Session:\n", NSub)
			fmt.Println("Last Session:\n", NLastSession)
			fmt.Println()

			// fmt.Println("----------- Overhead from co-locate -----------")
			// for i := 0; i < len(N); i++ {
			// 	initial[i] = C[i].Number * NSub[i].CPU
			// }
			// fmt.Println(initial)

			fmt.Println("----------- Nodes' <FINAL Score> for", P[j].Name, "-----------") // Pod index
			var fs [NodeNum]float64                                                       // Store final score
			var min float64                                                               // Store the value of the max score

			for i := 0; i < len(N); i++ {
				FS[i].NodeName = NSub[i].Name
				FS[i].PodName = P[i].Name

				if Fail[i] { // assign a large delay value (9999999) to unavailable node(s)
					fs[i] = 9999999	// Should use something like INT_MAX, would be better
					FS[i].Score = 9999999
				} else {
					if NSub[i].CPU > NLastSession[i].CPU { // if the "computation delay" is larger in current session than the last session
						FS[i].Score = ((C[i].Number) * alpha * (NSub[i].CPU - NLastSession[i].CPU)) + (alpha * NSub[i].CPU) + (1-alpha)*D[i].Delay*2 // OBJECTIVE FUNCTION
						//            Calculating the overhead of co-locating pods v.s. deploy on other nodes
						//            Because the object function is minimizing the average delay from all requests
						fs[i] = FS[i].Score
					} else {
						FS[i].Score = alpha*NSub[i].CPU + (1-alpha)*D[i].Delay*2 // OBJECTIVE FUNCTION
						//            Calculating the overhead of co-locating pods v.s. deploy on other nodes
						//            Because the object function is minimizing the average delay from all requests
						fs[i] = FS[i].Score
					}
				}
			}
			// fmt.Println("Alpha =", alpha, "\t( Load weight =", alpha, ", delay weight =", 1-alpha, ")")
			fmt.Println("Node avg RTT:\t", D)
			fmt.Println("FINAL Latency:\t", FS)

			min = Min(fs)
			// var bind [PodNum]Binding // Store binding information Node/Pod

			count := 0
			for i := 0; i < len(N); i++ {
				if FS[i].Score == min { // find node with least latency

					count++ // indicate that a node with best latency has been found once

					bind[j].NodeName = FS[i].NodeName // Pod index
					bind[j].PodName = P[j].Name       // Pod index

					//fmt.Println("Best Node:\t", FS[i].NodeName, "\tLatency:", min)
					//C[i].Number += 1.0
				}
			}

			if count > 1 { // exsist node with the exact same BEST latency
				var mem [NodeNum]float64 // store current RAM on each node

				for i := 0; i < len(N); i++ {
					if FS[i].Score == min { // Store RAM values from the BEST nodes
						mem[i] = N[i].RAM
					} else { // ignore nodes other than best nodes
						mem[i] = 0
					}
				}

				bestMem := 0.0
				bestMem = Max(mem) // store the largest RAM value within BEST nodes

				for i := 0; i < len(N); i++ {
					if N[i].RAM == bestMem { // ignore nodes other than best nodes
						bind[j].NodeName = FS[i].NodeName // Pod index
						bind[j].PodName = P[i].Name       // Pod index
						fmt.Println("Best Node:\t", bind[j].NodeName, "\tLatency:", min*2)
						C[i].Number += 1.0
					}
				}
			} else {
				fmt.Println("Best Node:\t", bind[j].NodeName, "\tLatency:", min*2)
				for i := 0; i < len(N); i++ {
					if bind[j].NodeName == N[i].Name {
						C[i].Number += 1.0
					}
				}
			}

			fmt.Println("Bindind Result:\t", bind[j]) // confirm binding result
			fmt.Println()

			fmt.Println("Node resouce before deployment:\n", N)
			fmt.Println()

			for i := 0; i < len(N); i++ {
				if N[i].Name == bind[j].NodeName {
					N[i].CPU = N[i].CPU - P[j].CPU
					N[i].RAM = N[i].RAM - P[j].RAM
					N[i].B = N[i].B - P[j].B
				}
			}
			fmt.Println("Node resouce after deployment:\n", N, "\n- - - - - - E N D - - - - - -\n\n\n")
		}
	}
	fmt.Println("Deployment Result:\n", bind)
	fmt.Println()
	fmt.Println("Numbers of Pod(s) on each node:\n", C)
	fmt.Println()
	fmt.Println("Node status:\n", N)
	fmt.Println()
}
