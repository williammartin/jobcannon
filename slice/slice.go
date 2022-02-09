package slice

func Map[A any, B any](mapFn func(A) B, as []A) []B {
	mappedVals := []B{}
	for _, a := range as {
		mappedVals = append(mappedVals, mapFn(a))
	}
	return mappedVals
}

func Drop[A any](predicateFn func(A) bool, as []A) []A {
	retainedValues := []A{}
	for _, a := range as {
		if !predicateFn(a) {
			retainedValues = append(retainedValues, a)
		}
	}
	return retainedValues
}
