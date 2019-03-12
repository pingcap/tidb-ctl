// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/base64"
	"fmt"
)

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func askForConfirmation() (bool, error) {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}
	okResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	noResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okResponses, response) {
		return true, nil
	} else if containsString(noResponses, response) {
		return false, nil
	} else {
		fmt.Println("Please type yes or no and then press enter")
		return askForConfirmation()
	}
}

func containsString(slice []string, element string) bool {
	for _, ele := range slice {
		if ele == element {
			return true
		}
	}
	return false
}
