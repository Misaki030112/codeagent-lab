package calc

func Add(a, b int) int {
	return a + b
}

func Divide(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}
