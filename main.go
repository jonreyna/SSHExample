package main

import (
	"io/ioutil" // makes writing files easier
	"log"       // to log if there's an error
	"net"       // to join the IP address with port

	"golang.org/x/crypto/ssh" // for connecting via SSH
)

// Your login credentials
var clientConfig = ssh.ClientConfig{
	User: "username",
	Auth: []ssh.AuthMethod{
		ssh.Password("password"),
	},
}

func main() {

	// There's only one CLICommand object per IP address
	type CLICommand struct {
		On  string   // the IP to run this CLI command on
		Run []string // the commands to run on this IP
	}

	// Define the commands we're going to run, and on which boxes (or IPs)
	cliCommands := []CLICommand{
		{
			On: "192.168.1.1",
			Run: []string{
				"show interfaces detail",
				"show chassis hardware detail",
			},
		},
	}

	// Loop over every CLI command object, and run the defined commands
	for _, cmd := range cliCommands {
		output, err := ExecSSH(cmd.On, cmd.Run...)
		if err != nil {
			log.Printf("error %q executing %v on %q", err.Error(), cmd.Run, cmd.On)
			continue // if we have an error, we'll log it, but keep going
		}

		// write the output to a file with the IP address
		if err := ioutil.WriteFile(cmd.On+".txt", output, 0755); err != nil {
			log.Printf("error %q writing file %q", err, cmd.On+".txt")
		}
	}
}

func ExecSSH(ip string, commands ...string) ([]byte, error) {

	// open the SSH client session
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), &clientConfig)
	if err != nil {
		return nil, err
	}

	// make sure we close the client when we're done
	defer client.Close()

	// results will hold all the results for all commands
	var results []byte

	// loop over all the commands for this IP
	for _, cmd := range commands {

		// open a new session on the SSH client
		session, err := client.NewSession()
		if err != nil {
			// if there's an error, we quit executing all commands on this IP
			// close the session, and return the error
			session.Close()
			return nil, err
		}

		// combine stdout, and stderr output of the command execution into one byte slice
		allOut, err := session.CombinedOutput(cmd)
		if err != nil {
			session.Close()
			return nil, err
		}

		// close the session, so we don't leak memory and have a gajillion sessions open on the router
		session.Close()
		// append all the output results to the results byte slice we're going to return
		results = append(results, allOut...)
	}

	return results, nil
}
