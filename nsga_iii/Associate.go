package nsga_iii

import (
	"math"
)

func Associate(currentPopulation Population, referencePoints []*ReferencePoint) {
	numberOfObjectiveFunctions := len(currentPopulation[0].ObjectiveValues)
	for _, individual := range currentPopulation {
		individual.PerpendicularDistance = math.MaxFloat64
		for _, referencePoint := range referencePoints {
			perpendicularDistance := ComputePerpendicularDistance(individual.NormalizedObjectiveValues, *referencePoint, numberOfObjectiveFunctions)
			if individual.PerpendicularDistance > perpendicularDistance {
				individual.PerpendicularDistance = perpendicularDistance
				individual.ReferencePoint = *referencePoint
			}
		}

	}
}

func ComputePerpendicularDistance(normalizedObjectiveValues []float64, referencePoint ReferencePoint, numberOfObjectiveValues int) float64{
	numerator := 0.0
	denominator := 0.0
	for i := 0; i < numberOfObjectiveValues; i++{
		numerator += normalizedObjectiveValues[i]*referencePoint.Coordinates[i]
		denominator += math.Pow(referencePoint.Coordinates[i], 2.0)
	}

	ratio := numerator/denominator

	distance := 0.0

	for i := 0; i < numberOfObjectiveValues; i++{
		distance += math.Pow(referencePoint.Coordinates[i]*ratio - normalizedObjectiveValues[i], 2.0)
	}

	return math.Sqrt(distance)
}
