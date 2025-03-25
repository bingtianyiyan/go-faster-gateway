package utils

import (
	"testing"
)

func TestDeepCopy(t *testing.T) {
	type User struct {
		UserName string
		Address  string
		Hoby     []string
	}

	userdata := new(User)
	userdata.UserName = "jack"
	userdata.Address = "上海"

	userdata2 := new(User)
	userdata2.UserName = "jack2"
	userdata2.Hoby = []string{
		"Play", "Music",
	}

	newUserData, _ := DeepCopy(nil, userdata)
	newUserData, _ = DeepCopy(newUserData, userdata2)
}
