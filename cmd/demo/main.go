package main

import (
	"clean-arch-go/domain/errors"
	"fmt"
)

func main() {
	fmt.Println(func1().Error())
}

func func1() error {
	return errors.E(errors.Op("func1"), func2())
}

func func2() error {
	return errors.E(errors.Op("func2"), func3())
}

func func3() error {
	return errors.E(errors.Op("func3"), func4())
}

func func4() error {
	return errors.E(errors.Op("func4"), fmt.Errorf("error4"))
}
