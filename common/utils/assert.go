package utils

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustTrue(ok bool) {
	if !ok {
		panic("must true")
	}
}
