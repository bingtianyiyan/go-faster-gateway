package utils

import "encoding/json"

func JsonMarshalToStr(src interface{}) (string, error) {
	bt, err := json.Marshal(src)
	if err != nil {
		return "", err
	}
	return string(bt), nil
}

func JsonMarshalToStrNoErr(src interface{}) string {
	bt, err := json.Marshal(src)
	if err != nil {
		return ""
	}
	return string(bt)
}
