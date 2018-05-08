package activedir

import (
	"fmt"
	"os"
	"strings"

	"github.com/mavricknz/ldap"
)

// GetGroupsForUser makes an LDAP connection to find the groups in Active Directory.
func GetGroupsForUser(userID string) ([]string, error) {
	groups := []string{}
	conn := ldap.NewLDAPConnection(
		"cad3.byu.edu",
		3268)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	username := os.Getenv("LDAP_USERNAME")
	password := os.Getenv("LDAP_PASSWORD")
	err = conn.Bind(username, password)
	if err != nil {
		panic(err)
	}
	search := ldap.NewSearchRequest(
		"ou=people,dc=byu,dc=local",
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,
		0,
		false,
		fmt.Sprintf("(Name=%s)", userID),
		[]string{"Name", "MemberOf"},
		nil,
	)
	res, err := conn.Search(search)
	if err != nil {
		panic(err)
	}
	//verify name
	for i := 0; i < len(res.Entries); i++ {
		name := res.Entries[i].GetAttributeValue("Name")
		if name != userID {
			continue
		}

		groupsTemp := res.Entries[0].GetAttributeValues("MemberOf")
		groups = translateGroups(groupsTemp)
	}

	return groups, nil
}

func translateGroups(groups []string) []string {
	toReturn := []string{}

	for _, entry := range groups {
		AD := strings.Split(entry, ",")
		toReturn = append(toReturn, strings.TrimPrefix(AD[0], "CN="))
	}
	return toReturn
}
