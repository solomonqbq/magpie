package task

import (
	"fmt"
	"testing"
)

func TestUrl(t *testing.T) {
	url, err := Parse("abc://shi:gx@sdfsfs?xxx=u&yy=dsfwe")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(url)
}
