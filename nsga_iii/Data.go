package nsga_iii

import (
	"math"
)

type Resources struct {
	CpuCores float64
	Memory   float64
}

type Power struct {
	IdlePower float64
	MaxPower  float64
}

type Task struct {
	TaskID            string
	RequiredResources Resources
	NodeID            string
	TaskType          string
}

type Node struct {
	ID                 string
	AvailableResources Resources
	RemainingResources *Resources
	Tasks              map[string]Task
	Power              Power

	CpuWeight float64
	MemoryWeight float64
	CpuQuotient float64
	MemoryQuotient float64
}

//for Genetic Algorithm
type Individual struct {
	ID                               string
	NodeIdOfTaskIdAssignment         map[string]string
	NodeIdOfTaskIdOriginalAssignment map[string]string
	AllNodes                         map[string]Node
	AllTasks                         map[string]Task
	NumberOfUnassignedTasks          int
	NumberOfUnlessNodes              int

			ObjectiveValues           []float64
	TranslatedObjectiveValues []float64
	NormalizedObjectiveValues []float64
	PerpendicularDistance     float64
	ReferencePoint            ReferencePoint

	//for NSGA-II
	IndividualsDominatedByThis      []*Individual
	NumberOfIndividualsDominateThis int
	Rank                            int
	CrowdingDistance                float64

	//NSGA III
	ConstrainedViolationValue float64
	IsFeasible bool
}

type ReferencePoint struct{
	ID string
	Coordinates []float64
	NicheCount int
}

func (individual *Individual) init(originalNodes []Node, originalTasks []Task) {
	individual.AllNodes = make(map[string]Node)
	for _, node := range originalNodes {
		node.RemainingResources = &Resources{}
		individual.AllNodes[node.ID] = node
	}

	individual.AllTasks = make(map[string]Task)
	for _, task := range originalTasks {
		individual.AllTasks[task.TaskID] = task
	}

	individual.ComputeValues()
	individual.TranslatedObjectiveValues = make([]float64, len(individual.ObjectiveValues))
	individual.NormalizedObjectiveValues = make([]float64, len(individual.ObjectiveValues))

}


func (individual *Individual) ComputeValues(){
	individual.computeRemainingResources()
	individual.computeUnassignedTasks()
	individual.computeNumberOfUselessNodes()
	individual.computeObjectiveFunctions()
	individual.ConstrainedViolationValue = individual.ComputeConstrainedViolationValue()
	individual.IsFeasible = individual.CheckIsFeasible()
}
func (individual *Individual) computeRemainingResources() {

	for _, node := range individual.AllNodes{
		for task := range node.Tasks{
			delete(node.Tasks, task)
		}
		node.RemainingResources.Memory = node.AvailableResources.Memory
		node.RemainingResources.CpuCores = node.AvailableResources.CpuCores
	}

	counter := 0
	for taskID, _ := range individual.NodeIdOfTaskIdAssignment {
		counter++
		nodeID := individual.NodeIdOfTaskIdAssignment[taskID]
		if len(nodeID) != 0 {
			task := individual.AllTasks[taskID]
			node := individual.AllNodes[nodeID]
			task.NodeID = node.ID
			if len(node.Tasks) == 0 {
				node.Tasks = make(map[string]Task)
			}
			node.Tasks[task.TaskID] = task
			node.RemainingResources.CpuCores = node.RemainingResources.CpuCores - task.RequiredResources.CpuCores
			node.RemainingResources.Memory = node.RemainingResources.Memory - task.RequiredResources.Memory

			individual.AllTasks[taskID] = task
			individual.AllNodes[nodeID] = node
		}
	}

}

func (individual *Individual) computeNumberOfUselessNodes(){
	individual.NumberOfUnlessNodes = 0
	for _, node := range individual.AllNodes{
		if len(node.Tasks) == 0{
			individual.NumberOfUnlessNodes++
		}
	}
}
func (individual *Individual) computeUnassignedTasks() {
	individual.NumberOfUnassignedTasks = 0
	for _, nodeID := range individual.NodeIdOfTaskIdAssignment {
		if len(nodeID) == 0 {
			//this task is not assigned to node
			individual.NumberOfUnassignedTasks++
		}
	}
}

func (individual *Individual) computeObjectiveFunctions() {
	individual.ObjectiveValues = []float64{}
	individual.ObjectiveValues = append(individual.ObjectiveValues, float64(individual.computeSpreadObjectiveFunction()))
	individual.ObjectiveValues = append(individual.ObjectiveValues, float64(individual.computeUniquenessObjectiveFunction()))
	individual.ObjectiveValues = append(individual.ObjectiveValues, individual.computePowerObjectiveFunction())
	individual.ObjectiveValues = append(individual.ObjectiveValues, individual.computeResourcesUtilizationObjectiveFunction())
}

func (individual *Individual) computeSpreadObjectiveFunction() int {
	spreadObjectiveValue := 0

	for _, node := range individual.AllNodes {
		nodeSpreadObjectiveValue := 0
		for i := 0; i < len(node.Tasks); i++ {
			nodeSpreadObjectiveValue += i + 1
		}
		spreadObjectiveValue += nodeSpreadObjectiveValue
	}

	return spreadObjectiveValue
}

func (individual *Individual) computeUniquenessObjectiveFunction() int {
	totalUniquenessObjectiveValue := 0
	for _, node := range individual.AllNodes {
		replicasOfTaskType := make(map[string]int)
		for _, task := range node.Tasks {
			replicasOfTaskType[task.TaskType] += 1
		}
		nodeUniquenessObjectiveValue := 0
		for _, replicas := range replicasOfTaskType {
			for i := 0; i < replicas; i++ {
				nodeUniquenessObjectiveValue += i + 1
			}
		}
		totalUniquenessObjectiveValue += nodeUniquenessObjectiveValue
	}
	return totalUniquenessObjectiveValue
}

func (individual *Individual) computePowerObjectiveFunction() float64 {
	totalPower := 0.0
	for _, node := range individual.AllNodes {
		totalPower += (node.Power.MaxPower -node.Power.IdlePower)*
			((node.AvailableResources.Memory -node.RemainingResources.Memory)/
				node.AvailableResources.Memory) + node.Power.IdlePower
	}
	return totalPower
}

func (individual *Individual) computeAssignmentDifferenceObjectiveFunction() int {
	if len(individual.NodeIdOfTaskIdOriginalAssignment) == 0 {
		return 0
	}
	differenceObjectiveValue := 0
	for taskIdOriginal, nodeIdOriginal := range individual.NodeIdOfTaskIdOriginalAssignment {
		if individual.NodeIdOfTaskIdAssignment[taskIdOriginal] != nodeIdOriginal {
			differenceObjectiveValue++
		}
	}

	return differenceObjectiveValue
}

func (individual *Individual) computeMemoryUtilizationObjectiveFunction() float64 {
	totalOverResourceUtilization := 0.0
	for _, node := range individual.AllNodes {
		if node.RemainingResources.Memory > 0 {
			totalOverResourceUtilization += math.Abs(node.RemainingResources.Memory)
		}
	}
	return totalOverResourceUtilization
}

func (individual *Individual) computeResourcesUtilizationObjectiveFunction() float64{
	resourcesUtilizationObjectiveValue := 0.0
	for _, node := range individual.AllNodes{
		resourcesUtilizationObjectiveValue += math.Abs((node.RemainingResources.CpuCores/node.AvailableResources.CpuCores)-
			(node.RemainingResources.Memory/node.AvailableResources.Memory))
	}
	return resourcesUtilizationObjectiveValue/float64(len(individual.AllNodes))
}


func (individual *Individual) computeCPUUtilizationObjectiveFunction() float64 {
	totalOverResourceUtilization := 0.0
	for _, node := range individual.AllNodes {
		if node.RemainingResources.CpuCores > 0 {
			totalOverResourceUtilization += math.Abs(node.RemainingResources.CpuCores)
		}
	}
	return totalOverResourceUtilization
}

func (individual *Individual) computeResourceUtilizationObjectiveFunction2() float64 {
	totalOverResourceUtilization := 0.0
	for _, node := range individual.AllNodes {
		totalOverResourceUtilization += math.Abs(node.RemainingResources.Memory)
	}
	return totalOverResourceUtilization
}

func (individual *Individual) dominates(anotherIndividual Individual) bool {
	for i := 0; i < len(individual.ObjectiveValues); i++ {
		// no worse than
		if individual.ObjectiveValues[i] > anotherIndividual.ObjectiveValues[i] {
			return false
		}
	}
	for i := 0; i < len(individual.ObjectiveValues); i++ {
		// at least better than the other
		if individual.ObjectiveValues[i] < anotherIndividual.ObjectiveValues[i] {
			return true
		}
	}

	return false
}

func (individual *Individual) crowdedComparisonOperatorLess(anotherIndividual Individual) bool {
	return individual.Rank < anotherIndividual.Rank || ((individual.Rank == anotherIndividual.Rank) && (individual.CrowdingDistance > anotherIndividual.CrowdingDistance))
}


//---------------

func (individual *Individual) CheckIsFeasible() bool{
	for _, node := range individual.AllNodes{
		if node.RemainingResources.Memory < 0 || node.RemainingResources.CpuCores < 0{
			return false
		}
	}

	return true
}

func (individual *Individual) ComputeConstrainedViolationValue()float64{
	constrainedViolationValue := 0.0
	for _, node := range individual.AllNodes{
		if node.RemainingResources.Memory < 0 {
			constrainedViolationValue += math.Abs(node.RemainingResources.Memory)
		}

		if node.RemainingResources.CpuCores < 0 {
			constrainedViolationValue += math.Abs(node.RemainingResources.CpuCores)
		}
	}
	return constrainedViolationValue
}

func (individual *Individual) constraintDominate(anotherIndividual Individual)bool {
	if (individual.IsFeasible && !anotherIndividual.IsFeasible) ||
		(!individual.IsFeasible && !anotherIndividual.IsFeasible && individual.ConstrainedViolationValue < anotherIndividual.ConstrainedViolationValue) ||
		(individual.IsFeasible && anotherIndividual.IsFeasible && individual.dominates(anotherIndividual)) {
		return true
	} else {
		return false
	}
}