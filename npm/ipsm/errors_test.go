package ipsm

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-container-networking/npm/util"
)

/*
✅ | where Raw !contains "Set cannot be destroyed: it is in use by a kernel component" // Exit status 1
✅ | where Raw !contains "Elem separator in" // Error: There was an error running command: [ipset -A -exist azure-npm-527074092 10.104.7.252,3000] Stderr: [exit status 1, ipset v7.5: Syntax error: Elem separator in 10.104.7.252,3000, but settype hash:net supports none.]
✅ | where Raw !contains "The set with the given name does not exist" // Exit status 1
 | where Raw !contains "TSet cannot be created: set with the same name already exists" // Exit status 1
| where Raw !contains "failed to create ipset rules" // Exit status 1
| where Raw !contains "Second element is missing from" //Error: There was an error running command: [ipset -A -exist azure-npm-4041682038 172.16.48.55] Stderr: [exit status 1, ipset v7.5: Syntax error: Second element is missing from 172.16.48.55.]
| where Raw !contains "Error: failed to create ipset."
| where Raw !contains "Missing second mandatory argument to command del" //Error: There was an error running command: [ipset -D -exist azure-npm-2064349730] Stderr: [exit status 2, ipset v7.5: Missing second mandatory argument to command del Try `ipset help' for more information.]
| where Raw !contains "Kernel error received: maximal number of sets reached" //Error: There was an error running command: [ipset -N -exist azure-npm-804639716 nethash] Stderr: [exit status 1, ipset v7.5: Kernel error received: maximal number of sets reached, cannot create more.]
| where Raw !contains "failed to create ipset list"
| where Raw !contains "set with the same name already exists"
| where Raw !contains "failed to delete ipset entry"
| where Raw !contains "Set to be added/deleted/tested as element does not exist"
| where (Raw !contains "ipset list with" and Raw !contains "not found")
| where Raw !contains "cannot parse azure: resolving to IPv4 address failed" //Error: There was an error running command: [ipset -A -exist azure-npm-2711086373 azure-npm-3224564478] Stderr: [exit status 1, ipset v7.5: Syntax error: cannot parse azure: resolving to IPv4 address failed]
| where Raw !contains "Sets with list:set type cannot be added to the set" //Error: There was an error running command: [ipset -A -exist azure-npm-530439631 azure-npm-2711086373] Stderr: [exit status 1, ipset v7.5: Sets with list:set type cannot be added to the set.]
| where Raw !contains "failed to delete ipset" // Different from the one above
| where Raw !contains " The value of the CIDR parameter of the IP address" //Error: There was an error running command: [ipset -A -exist azure-npm-3142322720 10.2.1.227/0] Stderr: [exit status 1, ipset v7.5: The value of the CIDR parameter of the IP address is invalid]
| where (Raw !contains "Syntax error:" and Raw !contains "is out of range 0-32") // IPV6 addr Error: There was an error running command: [ipset -A -exist azure-npm-1389121076 fe80::/64] Stderr: [exit status 1, ipset v7.5: Syntax error: '64' is out of range 0-32]
| where Raw !contains "exit status 253" // need to understand
| where Raw !contains "Operation not permitted" // Most probably from UTs
*/

// "Set cannot be destroyed: it is in use by a kernel component"
func TestSetCannotBeDestroyed(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(util.IpsetTestConfigFile); err != nil {
		t.Fatalf("TestAddToList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(util.IpsetTestConfigFile); err != nil {
			t.Fatalf("TestAddToList failed @ ipsMgr.Restore")
		}
	}()

	if err := ipsMgr.CreateSet("test-set", append([]string{util.IpsetNetHashFlag})); err != nil {
		t.Fatalf("Failed to create set with err %v", err)
	}

	if err := ipsMgr.AddToSet("test-set", fmt.Sprintf("%s", "1.1.1.1"), util.IpsetIPPortHashFlag, "0"); err != nil {
		t.Fatalf("Failed to add to set with err %v", err)
	}

	if err := ipsMgr.AddToList("test-list", "test-set"); err != nil {
		t.Fatalf("Failed to add to list with err %v", err)
	}

	// Delete set and validate set is not exist.
	if err := ipsMgr.DeleteSet("test-set"); err != nil {
		if err.ErrID != SetCannotBeDestroyedInUseByKernelComponent {
			t.Fatalf("Expected to error with ipset in use by kernel component")
		}
	}
}

func TestElemSeparatorSupportsNone(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(util.IpsetTestConfigFile); err != nil {
		t.Fatalf("TestAddToList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(util.IpsetTestConfigFile); err != nil {
			t.Fatalf("TestAddToList failed @ ipsMgr.Restore")
		}
	}()

	if err := ipsMgr.CreateSet("test-set", append([]string{util.IpsetNetHashFlag})); err != nil {
		t.Fatalf("TestAddToList failed @ ipsMgr.CreateSet")
	}

	entry := &ipsEntry{
		operationFlag: util.IpsetTestFlag,
		set:           util.GetHashedName("test-set"),
		spec:          append([]string{fmt.Sprintf("10.104.7.252,3000")}),
	}

	if _, err := ipsMgr.Run(entry); err == nil {
		t.Errorf("")
	}
}

func TestIPSetWithGivenNameDoesNotExist(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(util.IpsetTestConfigFile); err != nil {
		t.Fatalf("TestAddToList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(util.IpsetTestConfigFile); err != nil {
			t.Fatalf("TestAddToList failed @ ipsMgr.Restore")
		}
	}()

	entry := &ipsEntry{
		operationFlag: util.IpsetAppendFlag,
		set:           util.GetHashedName("test-set"),
		spec:          append([]string{util.GetHashedName("test-set2")}),
	}

	var err *NPMError
	if _, err = ipsMgr.Run(entry); err == nil {
		t.Fatalf("Expected set to not exist when adding to nonexistent set")
	}

	if err.ErrID != SetWithGivenNameDoesNotExist {
		t.Fatalf("Expected error code to match when set does not exist")
	}
}

func TestIPSetWithGivenNameAlreadyExists(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(util.IpsetTestConfigFile); err != nil {
		t.Fatalf("TestAddToList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(util.IpsetTestConfigFile); err != nil {
			t.Fatalf("TestAddToList failed @ ipsMgr.Restore")
		}
	}()

	var err *NPMError
	entry := &ipsEntry{
		name:          "test-set",
		operationFlag: util.IpsetCreationFlag,
		// Use hashed string for set name to avoid string length limit of ipset.
		set:  util.GetHashedName("test-set"),
		spec: append([]string{util.IpsetNetHashFlag}),
	}

	if errCode, err := ipsMgr.Run(entry); err != nil && errCode != 1 {
		t.Fatalf("Expected err")
	}

	if errCode, err := ipsMgr.Run(entry); err != nil && errCode != 1 {

		t.Fatalf("Expected err")
	}

	if err.ErrID != SetWithGivenNameDoesNotExist {
		t.Fatalf("Expected error code to match when set does not exist")
	}
}
