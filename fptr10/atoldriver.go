package main

import (
	fptr10 "atoldriver/fptr"
	"fmt"
)

func main() {
	fptr, err := fptr10.NewSafe()
	//fptr, err := fptr10.NewWithPath("C:\\Program Files\\ATOL\\Drivers10\\KKT\\bin")
	if err != nil {
		fmt.Println("Ошибка")
		fmt.Println(err)
		return
	}
	fmt.Println(fptr.Version())
	fptr.Destroy()
}

func SendComandeAndGetAnswerFromKKT(fptr *fptr10.IFptr, comJson string) (string, error) {
	fptr.setParam(fptr.LIBFPTR_PARAM_JSON_DATA, comJson)
	fptr.processJson()
	result := fptr.GetParamString(fptr10.LIBFPTR_PARAM_JSON_DATA)
	return result, nil
}
