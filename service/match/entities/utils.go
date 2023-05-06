// entities
// @author LanguageY++2013 2022/11/11 11:04
// @company soulgame
package entities

import (
	"strings"
)

//工具集

func inStringArray(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func inTeamNames(name string, teamNames string) bool {
	if teamNames == "*" {
		return true
	}

	return inStringArray(name, strings.Split(teamNames, ","))
}
