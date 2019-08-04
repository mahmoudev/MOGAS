package nsga_iii

import (
	"fmt"
	"github.com/rs/xid"
	"strconv"
)

var nsga2 NSGA2

type NSGA3 struct {
}

func (nsga3 NSGA3) GenerateNextPopulation(t int, g GeneticAlgorithm, parentPopulation Population, referencePoints []*ReferencePoint) Population {
	//output P(t+1)
	var nextPopulation Population
	//[ALGORITHM-1]STEP-1
	var temporaryNextPopulation Population //S
	i := 0
	//[ALGORITHM-1]STEP-2
	newPopulation := g.makeNewPopulation(parentPopulation, true)
	//[ALGORITHM-1]STEP-3
	unionOfParentAndNewPopulations := g.combinePopulation(parentPopulation, newPopulation)

	//[ALGORITHM-1]STEP-4
	fronts := nsga2.performFastNonDominatedSort(unionOfParentAndNewPopulations)
	//[ALGORITHM-1]STEP-5,6,7

	for ok := true; ok; ok = len(temporaryNextPopulation) <= g.PopulationSize {
		nsga3.unionFrontWithTemporaryPopulation(&temporaryNextPopulation, *fronts[i])
		i++
	}

	//[ALGORITHM-1]STEP-8
	lastFront := fronts[i-1]
	//[ALGORITHM-1]STEP-9
	if len(temporaryNextPopulation) == g.PopulationSize {
		//[ALGORITHM-1]STEP-10
		nextPopulation = temporaryNextPopulation
	} else {
		//[ALGORITHM-1]STEP-12
		nextPopulation = nsga3.unionFrontsUntilLevel(fronts, i-1)
		//[ALGORITHM-1]STEP-13
		numberOfRemainingIndividuals := g.PopulationSize - len(nextPopulation)
		//[ALGORITHM-1]STEP-14
		Normalize(temporaryNextPopulation)
		//[ALGORITHM-1]STEP-15
		Associate(temporaryNextPopulation, referencePoints)
		//[ALGORITHM-1]STEP-16
		nsga3.computeNicheCountForEachReferencePoint(nextPopulation, referencePoints)
		//fmt.Println("")
		//[ALGORITHM-1]STEP-17
		Niching(numberOfRemainingIndividuals, &temporaryNextPopulation, referencePoints, lastFront, &nextPopulation)
	}
	return nextPopulation
}

func (nsga3 NSGA3) unionFrontWithTemporaryPopulation(temporaryPopulation *Population, front Front) {
	for _, individual := range front {
		*temporaryPopulation = append(*temporaryPopulation, individual)
	}
}

func (nsga3 NSGA3) unionFrontsUntilLevel(fronts Fronts, level int) Population {
	var nextPopulation Population
	for i := 0; i < level; i++ {
		for _, individual := range *fronts[i] {
			nextPopulation = append(nextPopulation, individual)
		}
	}
	return nextPopulation
}

func (nsga3 NSGA3) computeNicheCountForEachReferencePoint(temporaryPopulation Population, referencePoints []*ReferencePoint) {
	for _, individual := range temporaryPopulation {
		for _, referencePoint := range referencePoints {
			if individual.ReferencePoint.ID == referencePoint.ID {
				referencePoint.NicheCount++
			}
		}
	}
}

var referencePointsCoordinates map[string]ReferencePoint

func (nsga3 NSGA3) GetReferencePoints(numberOfObjectiveFunctions int, numberOfSegments int) []*ReferencePoint {
	referencePointsCoordinates = map[string]ReferencePoint{}
	initialReferencePointCoordination := make([]float64, numberOfObjectiveFunctions)
	initialReferencePointCoordination[0] = 1.0
	initialReferencePoint := ReferencePoint{Coordinates: initialReferencePointCoordination}
	nsga3.generateReferencePointCoordinatesRecursively(initialReferencePoint, numberOfObjectiveFunctions, numberOfSegments)

	var referencePointsToReturn []*ReferencePoint
	for _, referencePointCoordinate := range referencePointsCoordinates {
		guid := xid.New()
		referencePointsToReturn = append(referencePointsToReturn, &ReferencePoint{ID: guid.String(), Coordinates: referencePointCoordinate.Coordinates})
	}

	return referencePointsToReturn
}
func (nsga3 NSGA3) generateReferencePointCoordinatesRecursively(referencePoint ReferencePoint, numberOfObjectiveFunctions int, numberOfSegments int) int {
	if referencePoint.Coordinates[0] < 0 {
		return 0
	}
	referencePointsCoordinates[fmt.Sprint(referencePoint.Coordinates)] = referencePoint

	for i := 1; i < numberOfObjectiveFunctions; i++ {
		newCoordinates := make([]float64, len(referencePoint.Coordinates))
		copy(newCoordinates, referencePoint.Coordinates)
		newCoordinates[0] = newCoordinates[0] - nsga3.Round(1.0/float64(numberOfSegments))
		newCoordinates[i] = newCoordinates[i] + nsga3.Round(1.0/float64(numberOfSegments))
		if _, exists := referencePointsCoordinates[fmt.Sprint(newCoordinates)]; !exists {
			newReferencePoint := ReferencePoint{Coordinates: newCoordinates}
			nsga3.generateReferencePointCoordinatesRecursively(newReferencePoint, numberOfObjectiveFunctions, numberOfSegments)
		}
	}

	return 0

}

func (nsga3 NSGA3) Round(number float64) float64 {
	float, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", number), 64)
	return float
}