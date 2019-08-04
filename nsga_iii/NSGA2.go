package nsga_iii

import (
	"math"
	"sort"
)

type Front []*Individual
type Fronts []*Front

type NSGA2 struct{}

func (n NSGA2) performFastNonDominatedSort(population Population) Fronts {
	var fronts Fronts
	for _, individual := range population {
		for _, anotherIndividual := range population {
			if individual.ID != anotherIndividual.ID{
				if individual.constraintDominate(*anotherIndividual) {
					individual.IndividualsDominatedByThis = append(individual.IndividualsDominatedByThis, anotherIndividual)
				} else if anotherIndividual.constraintDominate(*individual) {
					individual.NumberOfIndividualsDominateThis++
				}
			}
		}

		if individual.NumberOfIndividualsDominateThis == 0 {
			individual.Rank = 0
			if len(fronts) == 0 {
				fronts = append(fronts, &Front{})
			}
			*fronts[0] = append(*fronts[0], individual)
		}
	}

	frontCounter := 0
	for len(*fronts[frontCounter]) != 0 {
		var nextFront Front
		for _, individual := range *fronts[frontCounter] {
			for _, dominatedIndividual := range individual.IndividualsDominatedByThis {
				dominatedIndividual.NumberOfIndividualsDominateThis--
				if dominatedIndividual.NumberOfIndividualsDominateThis == 0 {
					dominatedIndividual.Rank = frontCounter + 1
					nextFront = append(nextFront, dominatedIndividual)
				}
			}
		}
		frontCounter++
		fronts = append(fronts, &nextFront)
	}

	return fronts
}

func (n NSGA2) computeCrowdingDistance(front Front) {
	if len(front) == 0 {
		return
	}

	for _, individual := range front {
		individual.CrowdingDistance = 0
	}
	//fmt.Println("fnt lenght " , len(front))
	for i := 0; i < len(front[0].ObjectiveValues); i++ {
		sort.Slice(front, func(indexOfFirst, indexOfSecond int) bool {
			return front[indexOfFirst].ObjectiveValues[i] > front[indexOfSecond].ObjectiveValues[i]
		})
		front[0].CrowdingDistance = math.MaxFloat64
		front[len(front)-1].CrowdingDistance = math.MaxFloat64

		for j := 1; j < len(front)-1; j++ {
			front[j].CrowdingDistance = front[j].CrowdingDistance + ((front[j+1].ObjectiveValues[i] - front[j-1].ObjectiveValues[i]) /
				(front[0].ObjectiveValues[i] - front[len(front)-1].ObjectiveValues[i]))
		}

	}

}

func (n NSGA2) sortFront(front Front) {
	sort.Slice(front, func(i, j int) bool {
		return !front[i].crowdedComparisonOperatorLess(*front[j])
	})
}
