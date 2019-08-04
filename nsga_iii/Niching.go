package nsga_iii

import (
	"math"
	"math/rand"
)

func Niching(numberOfRemainingIndividuals int, temporaryPopulation *Population, referencePoints []*ReferencePoint, lastFront *Front, incompleteNextPopulation *Population){
	k := 0
	sum := 0
	for _, ref := range referencePoints{
		sum += ref.NicheCount
	}

	for k<numberOfRemainingIndividuals{
		referencePointsMinNicheCount := findReferencePointsWithMinNicheCount(referencePoints)
		randomReferencePointWithMinNicheCount := referencePointsMinNicheCount[rand.Intn(len(referencePointsMinNicheCount))]
		individualsBelongToMinReferencePointAndLastFront := findIndividualsBelongToMinReferencePointAndLastFront(temporaryPopulation, *lastFront, *randomReferencePointWithMinNicheCount)

		if len(individualsBelongToMinReferencePointAndLastFront) != 0{
			var individual *Individual
			if randomReferencePointWithMinNicheCount.NicheCount == 0{
				individual = getIndividualWithMinPerpendicularDistance(individualsBelongToMinReferencePointAndLastFront)
				*incompleteNextPopulation = append(*incompleteNextPopulation, individual)
			}else{
				individual = individualsBelongToMinReferencePointAndLastFront[rand.Intn(len(individualsBelongToMinReferencePointAndLastFront))]
				*incompleteNextPopulation = append(*incompleteNextPopulation, individual)
			}
			randomReferencePointWithMinNicheCount.NicheCount++
			removeIndividualFromLastFront(*individual, lastFront)
			k++
		}else{
			removeReferencePointWithMinNicheCountFromReferencePoints(*randomReferencePointWithMinNicheCount, &referencePoints)
		}
	}
}

func findReferencePointsWithMinNicheCount(referencePoints []*ReferencePoint) []*ReferencePoint{
	minNicheCount := math.MaxInt64
	for _, referencePoint := range referencePoints{
		if referencePoint.NicheCount < minNicheCount{
			minNicheCount = referencePoint.NicheCount
		}
	}

	var referencePointsThatHasMinNicheCount []*ReferencePoint
	for _, referencePoint := range referencePoints{
		if referencePoint.NicheCount == minNicheCount{
			referencePointsThatHasMinNicheCount = append(referencePointsThatHasMinNicheCount, referencePoint)
		}
	}
	return referencePointsThatHasMinNicheCount
}


func findIndividualsBelongToMinReferencePointAndLastFront(population *Population, lastFront Front, referencePointWithMinNicheCount ReferencePoint)[]*Individual{
	var individualsBelongToReferencePointWithMinNicheCount []*Individual
		for _, individualInLastFront := range lastFront{
			if individualInLastFront.ReferencePoint.ID == referencePointWithMinNicheCount.ID{
				individualsBelongToReferencePointWithMinNicheCount = append(individualsBelongToReferencePointWithMinNicheCount, individualInLastFront)
			}
		}
	return individualsBelongToReferencePointWithMinNicheCount
}

func getIndividualWithMinPerpendicularDistance(individualsBelongToMinReferencePointAndLastFront []*Individual)*Individual{
	individualWithMinPerpendicularDistance := individualsBelongToMinReferencePointAndLastFront[0]
	for _, individual := range individualsBelongToMinReferencePointAndLastFront {
		if individual.PerpendicularDistance < individualWithMinPerpendicularDistance.PerpendicularDistance{
			individualWithMinPerpendicularDistance = individual
		}
	}
	return individualWithMinPerpendicularDistance
}

func removeIndividualFromLastFront(individual Individual, lastFront *Front){
	indexOfIndividual := 0
	for i , individualInLastFront := range *lastFront{
		if individual.ID == individualInLastFront.ID{
			indexOfIndividual = i
		}
	}
	*lastFront = append((*lastFront)[:indexOfIndividual], (*lastFront)[indexOfIndividual+1:]...)
}

func removeReferencePointWithMinNicheCountFromReferencePoints(referencePointWithMinNicheCount ReferencePoint, referencePoints *[]*ReferencePoint){
	indexOfReferencePointToRemove := 0

	for i, referencePoint := range *referencePoints{
		if referencePoint.ID == referencePointWithMinNicheCount.ID{
			indexOfReferencePointToRemove = i
		}
	}
	*referencePoints = append((*referencePoints)[:indexOfReferencePointToRemove], (*referencePoints)[indexOfReferencePointToRemove+1:]...)
}
