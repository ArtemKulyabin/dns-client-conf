package dhclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

const (
	dhclientNameServers = "prepend domain-name-servers %s;"
)

func copyfile(src string, dst string) (err error) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)

	return err
}

func setNameServers(addrs []net.IP, configpath, configpathbackup string) (err error) {
	// Back up dhclient configuration file
	dhclientBackup, err := os.OpenFile(configpathbackup, os.O_RDWR, 0)
	if err != nil {
		if os.IsNotExist(err) == true {
			copyfile(configpath, configpathbackup)
		} else {
			return err
		}
	}
	defer dhclientBackup.Close()

	dhclientFile, err := os.OpenFile(configpath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer dhclientFile.Close()

	buf, err := ioutil.ReadAll(dhclientFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(buf), "\n")

	index := -1

	for i, line := range lines {
		if line == "# next line generated by dns-client-conf" {
			index = i
			break
		}
	}

	if index == -1 {
		index = len(lines) - 1
		newLines := make([]string, cap(lines)+3)
		copy(newLines, lines)
		lines = newLines
	}

	lines[index] = "# next line generated by dns-client-conf"

	if len(addrs) > 0 {

		var entries []string

		for _, addr := range addrs {
			entries = append(entries, addr.String())
		}

		lines[index+1] = fmt.Sprintf(dhclientNameServers, strings.Join(entries, ", "))

	} else {
		lines[index+1] = ""
	}

	lines[index+2] = "# prev line generated by dns-client-conf"

	err = ioutil.WriteFile(configpath, bytes.NewBufferString(strings.Join(lines, "\n")).Bytes(), 0)

	return err
}

// AddNameServers reads the list of ip addresses, dhclient daemon configuration
// file path. On first step backup configuration file if this do not exist.
// On second step add dns addresses to dhclient config.
func AddNameServers(addrs []net.IP, configpath, configpathbackup string) (err error) {
	/*for _, addr := range addrs {
		err = helpers.CheckIP(addr)
		if err != nil {
			return err
		}
	}*/

	err = setNameServers(addrs, configpath, configpathbackup)

	return err
}

// RemoveNameServers consume dhclient daemon configuration
// file path and revert previously called AddNameServers method effect.
func RemoveNameServers(configpath, configpathbackup string) (err error) {
	return setNameServers([]net.IP{}, configpath, configpathbackup)
}
