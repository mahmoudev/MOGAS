package nsga_iii

import (
	"math"
	"errors"
)

func Normalize(populationWithOverflow Population){
	numberOfObjectiveFunctions := len(populationWithOverflow[0].ObjectiveValues)
	idealObjectivePoint := make([]float64, numberOfObjectiveFunctions)
	extremeTranslatedObjectivePoints := make([][]float64, numberOfObjectiveFunctions)
	extremeObjectivePoints := make([][]float64, numberOfObjectiveFunctions)
	//[ALGORITHM-2]STEP-1
	for indexOfObjectiveFunction := 0; indexOfObjectiveFunction < numberOfObjectiveFunctions; indexOfObjectiveFunction++{
		//[ALGORITHM-2]STEP-2
		idealValueOfObjectiveFunction := computeIdealObjectiveValue(populationWithOverflow, indexOfObjectiveFunction)
		idealObjectivePoint[indexOfObjectiveFunction] = idealValueOfObjectiveFunction

		//[ALGORITHM-2]STEP-3
		translateObjectiveFunction(populationWithOverflow, indexOfObjectiveFunction, idealValueOfObjectiveFunction)

		//[ALGORITHM-2]STEP-4
	}


	for indexOfObjectiveFunction := 0; indexOfObjectiveFunction < numberOfObjectiveFunctions; indexOfObjectiveFunction++ {
		extremeTranslatedObjectivePoints[indexOfObjectiveFunction] = computeTranslatedExtremeObjectivePoint(numberOfObjectiveFunctions, indexOfObjectiveFunction, populationWithOverflow)
	}
	for indexOfObjectiveFunction := 0; indexOfObjectiveFunction < numberOfObjectiveFunctions; indexOfObjectiveFunction++ {
		extremeObjectivePoints[indexOfObjectiveFunction] = computeExtremeObjectivePoint(numberOfObjectiveFunctions, indexOfObjectiveFunction, populationWithOverflow)
	}

	intercepts := ComputeIntercepts(extremeTranslatedObjectivePoints, numberOfObjectiveFunctions)

	for _, individual := range populationWithOverflow {
		for i := 0; i < numberOfObjectiveFunctions; i++ {
			//************************************************************changed to intercept
			if math.Abs(intercepts[i] - idealObjectivePoint[i]) > 10e-10{
				individual.NormalizedObjectiveValues[i] = individual.TranslatedObjectiveValues[i]/ (intercepts[i] - idealObjectivePoint[i])
			}else{
				individual.NormalizedObjectiveValues[i] = individual.TranslatedObjectiveValues[i] / 10e-10
			}
		}
	}
}

func computeTranslatedExtremeObjectivePoint(numberOfObjectiveFunctions int, indexOfObjectiveFunction int, populationWithOverflow Population) []float64{
	randomTranslatedObjectiveValue := populationWithOverflow[0].TranslatedObjectiveValues
	var extremeObjectivePoint []float64
	weightVector := initWeightVector(numberOfObjectiveFunctions, indexOfObjectiveFunction)
	minimumASFValue := ASF(randomTranslatedObjectiveValue, weightVector)
	extremeObjectivePoint = randomTranslatedObjectiveValue

	for _, individual := range populationWithOverflow {
		individualASFValue := ASF(individual.TranslatedObjectiveValues, weightVector)
		if minimumASFValue > individualASFValue {
			minimumASFValue = individualASFValue
			extremeObjectivePoint = individual.TranslatedObjectiveValues
		}
	}
	return extremeObjectivePoint
}

func computeExtremeObjectivePoint(numberOfObjectiveFunctions int, indexOfObjectiveFunction int, populationWithOverflow Population) []float64{
	randomObjectiveValue := populationWithOverflow[0].ObjectiveValues
	var extremeObjectivePoint []float64
	weightVector := initWeightVector(numberOfObjectiveFunctions, indexOfObjectiveFunction)
	minimumASFValue := ASF(randomObjectiveValue, weightVector)
	extremeObjectivePoint = randomObjectiveValue

	for _, individual := range populationWithOverflow {
		individualASFValue := ASF(individual.ObjectiveValues, weightVector)
		if minimumASFValue > individualASFValue {
			minimumASFValue = individualASFValue
			extremeObjectivePoint = individual.ObjectiveValues
		}
	}
	return extremeObjectivePoint
}


func translateObjectiveFunction(populationWithOverflow Population, indexOfObjectiveFunction int, idealValueOfObjectiveFunction float64) {
	for _, individual := range populationWithOverflow {
		individual.TranslatedObjectiveValues[indexOfObjectiveFunction] = individual.ObjectiveValues[indexOfObjectiveFunction]- idealValueOfObjectiveFunction
	}
}
func computeIdealObjectiveValue(populationWithOverflow Population, indexOfObjectiveFunction int) float64 {
	idealValueOfObjectiveFunction := populationWithOverflow[0].ObjectiveValues[indexOfObjectiveFunction]
	for _, individual := range populationWithOverflow {
		idealValueOfObjectiveFunction = math.Min(individual.ObjectiveValues[indexOfObjectiveFunction], idealValueOfObjectiveFunction)
	}
	return idealValueOfObjectiveFunction
}


func initWeightVector(size int, indexOfCurrentObjectiveFunction int)[]float64{
	weightVector := make([]float64, size)
	for i := 0; i < size; i++{
		if indexOfCurrentObjectiveFunction == i{
			weightVector[i] = 1
		}else{
			weightVector[i] = math.Pow(10, -6)
		}
	}
	return weightVector
}


func ASF(translatedObjectiveValues []float64, weight []float64) float64{
	max := translatedObjectiveValues[0]/weight[0]
	for i, translatedObjectiveValue := range translatedObjectiveValues{
		if max < translatedObjectiveValue/weight[i]{
			max = translatedObjectiveValue/weight[i]
		}
	}
	return max
}

func ComputeIntercepts(extremeTranslatedObjectivePoints [][]float64, numberOfObjectiveFunctions int)[]float64{
	intercepts := make([]float64, numberOfObjectiveFunctions)

	B := make([]float64, numberOfObjectiveFunctions)
	for i := 0; i < numberOfObjectiveFunctions; i++{
		B[i] = 1.0
	}

	result, error := GaussPartial(extremeTranslatedObjectivePoints, B)

	isThereNegativeIntercept := false
	for _, number := range result{
		if number < 0{
			isThereNegativeIntercept = true
		}
	}

	if error == nil && !isThereNegativeIntercept {
		for i, number := range result {
			intercepts[i] = 1 / number
		}
		return intercepts
	}else{
		for i := 0; i < numberOfObjectiveFunctions; i++{
			intercepts[i] = extremeTranslatedObjectivePoints[i][i]
		}
		return intercepts
	}
}

//from open source project
func GaussPartial(a0 [][]float64, b0 []float64) ([]float64, error) {
	// make augmented matrix
	m := len(b0)
	a := make([][]float64, m)
	for i, ai := range a0 {
		row := make([]float64, m+1)
		copy(row, ai)
		row[m] = b0[i]
		a[i] = row
	}
	// WP algorithm from Gaussian elimination page
	// produces row-eschelon form
	for k := range a {
		// Find pivot for column k:
		iMax := k
		max := math.Abs(a[k][k])
		for i := k + 1; i < m; i++ {
			if abs := math.Abs(a[i][k]); abs > max {
				iMax = i
				max = abs
			}
		}
		if a[iMax][k] == 0 {
			return nil, errors.New("singular")
		}
		// swap rows(k, i_max)
		a[k], a[iMax] = a[iMax], a[k]
		// Do for all rows below pivot:
		for i := k + 1; i < m; i++ {
			// Do for all remaining elements in current row:
			for j := k + 1; j <= m; j++ {
				a[i][j] -= a[k][j] * (a[i][k] / a[k][k])
			}
			// Fill lower triangular matrix with zeros:
			a[i][k] = 0
		}
	}
	// end of WP algorithm.
	// now back substitute to get result.
	x := make([]float64, m)
	for i := m - 1; i >= 0; i-- {
		x[i] = a[i][m]
		for j := i + 1; j < m; j++ {
			x[i] -= a[i][j] * x[j]
		}
		x[i] /= a[i][i]
	}
	return x, nil
}