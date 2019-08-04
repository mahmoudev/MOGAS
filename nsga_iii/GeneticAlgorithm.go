package nsga_iii

import (
	"math/rand"
	"github.com/rs/xid"
	"fmt"
	"math"
	"time"
)

type GeneticAlgorithm struct {
	AllNodes                         []Node
	AllTasks                         []Task
	NodeIdOfTaskIdOriginalAssignment map[string]string
	PopulationSize                   int
	NumberOfGenerations              int
}

type Population []*Individual

func (g GeneticAlgorithm) GenerateRandomFeasibleIndividual() *Individual {
	nodes := make([]Node, len(g.AllNodes))
	for i, node := range g.AllNodes {
		nodes[i] = Node{RemainingResources: &Resources{Memory: node.AvailableResources.Memory, CpuCores: node.AvailableResources.CpuCores}, ID: node.ID}
	}

	shuffleNodes(nodes)

	guid := xid.New()
	nodeIdOfTaskIdAssignment := make(map[string]string)
	for _, task := range g.AllTasks {
		for _, node := range nodes {
			if task.RequiredResources.CpuCores <= node.RemainingResources.CpuCores &&
				task.RequiredResources.Memory <= node.RemainingResources.Memory {
				nodeIdOfTaskIdAssignment[task.TaskID] = node.ID
				node.RemainingResources.Memory -= task.RequiredResources.Memory
				node.RemainingResources.CpuCores -= task.RequiredResources.CpuCores
				break
			}
		}
	}

	newIndividual := Individual{ID: guid.String(), NodeIdOfTaskIdAssignment: nodeIdOfTaskIdAssignment, NodeIdOfTaskIdOriginalAssignment: g.NodeIdOfTaskIdOriginalAssignment}

	newIndividual.init(g.AllNodes, g.AllTasks)

	return &newIndividual

}

func shuffleNodes(nodes []Node) {
	rand.Seed(time.Now().UnixNano())
	for len(nodes) > 0 {
		n := len(nodes)
		randIndex := rand.Intn(n)
		nodes[n-1], nodes[randIndex] = nodes[randIndex], nodes[n-1]
		nodes = nodes[:n-1]
	}
}

func (g GeneticAlgorithm) GenerateRandomFeasiblePopulation() Population {
	randomPopulation := make([]*Individual, g.PopulationSize)
	for i := 0; i < g.PopulationSize; i++ {
		randomPopulation[i] = g.GenerateRandomFeasibleIndividual()
	}
	return randomPopulation
}
func (g GeneticAlgorithm) generateRandomIndividual() *Individual {
	guid := xid.New()
	nodeIdOfTaskIdAssignment := make(map[string]string)
	for _, task := range g.AllTasks {
		nodeIdOfTaskIdAssignment[task.TaskID] = g.AllNodes[rand.Intn(len(g.AllNodes))].ID
	}
	newIndividual := Individual{ID: guid.String(), NodeIdOfTaskIdAssignment: nodeIdOfTaskIdAssignment, NodeIdOfTaskIdOriginalAssignment: g.NodeIdOfTaskIdOriginalAssignment}
	newIndividual.init(g.AllNodes, g.AllTasks)
	return &newIndividual
}

func (g GeneticAlgorithm) generateRandomPopulation() Population {
	randomPopulation := make([]*Individual, g.PopulationSize)
	for i := 0; i < g.PopulationSize; i++ {
		randomPopulation[i] = g.generateRandomIndividual()
	}
	return randomPopulation
}

func (g GeneticAlgorithm) selectRandomIndividual(population Population) Individual {
	return *population[rand.Intn(g.PopulationSize)]
}

func (g GeneticAlgorithm) reproduce(firstIndividual Individual, secondIndividual Individual) Individual {
	newNodeIdOfTaskIdAssignment := make(map[string]string)
	numberOfTasks := len(firstIndividual.AllTasks)
	randomCut := rand.Intn(numberOfTasks)

	for i := 0; i < numberOfTasks; i++ {
		task := g.AllTasks[i]
		if i < randomCut {
			newNodeIdOfTaskIdAssignment[task.TaskID] = firstIndividual.NodeIdOfTaskIdAssignment[task.TaskID]
		} else {
			newNodeIdOfTaskIdAssignment[task.TaskID] = secondIndividual.NodeIdOfTaskIdAssignment[task.TaskID]
		}
	}
	guid := xid.New()
	newIndividual := Individual{ID: guid.String(), NodeIdOfTaskIdAssignment: newNodeIdOfTaskIdAssignment, NodeIdOfTaskIdOriginalAssignment: g.NodeIdOfTaskIdOriginalAssignment}
	newIndividual.init(g.AllNodes, g.AllTasks)
	newIndividual.ComputeValues()
	return newIndividual
}

func (g GeneticAlgorithm) mutate(individual *Individual) {
	swap := func(individual *Individual) {
		task1 := g.AllTasks[rand.Intn(len(g.AllTasks))]
		task2 := g.AllTasks[rand.Intn(len(g.AllTasks))]

		nodeID1 := individual.NodeIdOfTaskIdAssignment[task1.TaskID]
		nodeID2 := individual.NodeIdOfTaskIdAssignment[task2.TaskID]
		individual.NodeIdOfTaskIdAssignment[task1.TaskID] = nodeID2
		individual.NodeIdOfTaskIdAssignment[task2.TaskID] = nodeID1
	}
	change := func(individual *Individual) {
		task := g.AllTasks[rand.Intn(len(g.AllTasks))]
		node := g.AllNodes[rand.Intn(len(g.AllNodes))]
		individual.NodeIdOfTaskIdAssignment[task.TaskID] = node.ID
	}
	assignUnassigned := func(individual *Individual) {
		var unassignedTasks []string
		for taskID := range individual.NodeIdOfTaskIdAssignment {
			if individual.NodeIdOfTaskIdAssignment[taskID] == "" {
				unassignedTasks = append(unassignedTasks, taskID)
			}
		}
		unassignedTaskID := unassignedTasks[rand.Intn(len(unassignedTasks))]
		node := g.AllNodes[rand.Intn(len(g.AllNodes))]
		individual.NodeIdOfTaskIdAssignment[unassignedTaskID] = node.ID
	}
	unassignAssigned := func(individual *Individual) {
		var assignedTasks []string
		for taskID := range individual.NodeIdOfTaskIdAssignment {
			if individual.NodeIdOfTaskIdAssignment[taskID] != "" {
				assignedTasks = append(assignedTasks, taskID)
			}
		}
		assignedTaskID := assignedTasks[rand.Intn(len(assignedTasks))]
		individual.NodeIdOfTaskIdAssignment[assignedTaskID] = ""

	}

	probability := rand.Float64()
	if (probability <= 0.25) {
		change(individual)
	} else if (probability > 0.25 && probability <= 0.5) {
		swap(individual)
	} else if (probability > 0.5 && probability <= 0.51) {
		if(!individual.IsFeasible) {
			unassignAssigned(individual)
		}
	}else if(probability > 0.51 && probability <= 1.0){
		if(individual.NumberOfUnassignedTasks > 0) {
			assignUnassigned(individual)
		}
	}

	individual.ComputeValues()
}

func (g GeneticAlgorithm) combinePopulation(firstPopulation Population, secondPopulation Population) Population {
	combinedPopulation := make([]*Individual, g.PopulationSize*2)
	for i := 0; i < len(firstPopulation); i++ {
		combinedPopulation[i] = firstPopulation[i]
	}

	for i := len(secondPopulation); i < len(secondPopulation)*2; i++ {
		combinedPopulation[i] = secondPopulation[i-g.PopulationSize]
	}
	return combinedPopulation
}

func (g GeneticAlgorithm) binaryTormentSelection(population Population) Individual {
	firstIndividual := population[rand.Intn(g.PopulationSize)]
	secondIndividual := population[rand.Intn(g.PopulationSize)]

	if firstIndividual.crowdedComparisonOperatorLess(*secondIndividual) {
		return *firstIndividual
	} else {
		return *secondIndividual
	}
}

//constrained nsga iii
func (g GeneticAlgorithm) constrainedBinaryTournamentSelection(population Population) Individual {
	firstIndividual := population[rand.Intn(g.PopulationSize/2)]
	secondIndividual := population[g.PopulationSize/2+rand.Intn(g.PopulationSize/2)]

	if firstIndividual.IsFeasible && !secondIndividual.IsFeasible {
		return *firstIndividual
	} else if !firstIndividual.IsFeasible && secondIndividual.IsFeasible {
		return *secondIndividual
	} else if !firstIndividual.IsFeasible && !secondIndividual.IsFeasible {
		if firstIndividual.ConstrainedViolationValue > secondIndividual.ConstrainedViolationValue {
			return *secondIndividual
		} else if firstIndividual.ConstrainedViolationValue < secondIndividual.ConstrainedViolationValue {
			return *firstIndividual
		} else {
			if rand.Float64() > 0.5 {
				return *firstIndividual
			} else {
				return *secondIndividual
			}
		}
	} else {
		if rand.Float64() > 0.5 {
			return *firstIndividual
		} else {
			return *secondIndividual
		}
	}
}

func (g GeneticAlgorithm) makeNewPopulation(parentPopulation Population, isInitial bool) Population {
	newPopulation := make([]*Individual, g.PopulationSize)
	var firstIndividual Individual
	var secondIndividual Individual
	for i := 0; i < g.PopulationSize; i++ {
		firstIndividual = g.constrainedBinaryTournamentSelection(parentPopulation)
		secondIndividual = g.constrainedBinaryTournamentSelection(parentPopulation)
		newIndividual := g.reproduce(firstIndividual, secondIndividual)

		if rand.Float64()>0.5 {
			g.mutate(&newIndividual)
		}
		newPopulation [i] = &newIndividual
	}
	return newPopulation
}

func (g GeneticAlgorithm) RunGeneticAlgorithmNSGA2(numberOfSegments int) Population {
	nsga3 := NSGA3{}
	parentPopulation := g.GenerateRandomFeasiblePopulation()

	for t := 0; t < g.NumberOfGenerations; t++ {
		referencePoints := nsga3.GetReferencePoints(len(parentPopulation[0].ObjectiveValues), numberOfSegments)
		nextPopulation := nsga3.GenerateNextPopulation(t, g, parentPopulation, referencePoints)
		parentPopulation = nextPopulation
		for _, individual := range parentPopulation {
			individual.ReferencePoint = ReferencePoint{}
			individual.PerpendicularDistance = 0
			individual.IndividualsDominatedByThis = []*Individual{}
			individual.NumberOfIndividualsDominateThis = 0
			individual.Rank = 0
		}

	}
	return parentPopulation
}

func printRemainingResources(parentPopulation Population) {
	for _, n := range parentPopulation[0].AllNodes {
		fmt.Print("Node[", n.ID+"]:  ")
		for _, t := range n.Tasks {
			fmt.Print(t.TaskID + ", ")
		}
		fmt.Println("remainng ", n.RemainingResources, "cpu", (n.RemainingResources.CpuCores / n.AvailableResources.CpuCores),
			"mem", (n.RemainingResources.Memory / n.AvailableResources.Memory))
	}

	fmt.Println("+++++++++++++++++++++++")
	for _, task := range parentPopulation[0].AllTasks {
		if task.NodeID == "" {
			fmt.Println(task.TaskID, " ", task.RequiredResources)
		}
	}
}

func printAvg(parentPopulation Population, g GeneticAlgorithm, t int) {
	avgObjective := make([]float64, len(parentPopulation[0].ObjectiveValues))
	for _, ind := range parentPopulation {
		for i, obj := range ind.ObjectiveValues {
			avgObjective[i] += obj
		}
	}
	for i := range avgObjective {
		avgObjective[i] = avgObjective[i] / float64(g.PopulationSize)
	}
	fmt.Print(t, " ")
	for _, obj := range avgObjective {
		fmt.Print(obj, " ")
	}
	fmt.Println()
}

func printBest(parentPopulation Population, g GeneticAlgorithm, t int) {
	bestObjective := make([]float64, len(parentPopulation[0].ObjectiveValues))

	for i := range bestObjective {
		bestObjective[i] = math.MaxFloat64
	}
	for _, ind := range parentPopulation {
		for i, obj := range ind.ObjectiveValues {
			if obj < bestObjective[i] {
				bestObjective[i] = obj
			}
		}
	}

	fmt.Print(t, " ")
	for _, obj := range bestObjective {
		fmt.Print(obj, " ")
	}
	fmt.Println()
}

func printResourcesUtilization(parentPopulation Population) {
	for _, ind := range parentPopulation {
		for _, obj := range ind.ObjectiveValues {
			fmt.Print(obj, " ")
		}
		fmt.Println()
	}

}

func printResourcePercentageForEachNode(p Population) {
	var bestInd Individual
	bestVal := math.MaxFloat64
	for _, ind := range p {
		if ind.ObjectiveValues[2] < bestVal {
			bestVal = ind.ObjectiveValues[2]
			bestInd = *ind
		}
	}

	for _, node := range bestInd.AllNodes {
		fmt.Println(node.RemainingResources.CpuCores/node.AvailableResources.CpuCores,
			node.RemainingResources.Memory/node.AvailableResources.Memory)
	}
}
