package tunnel

import (
	"github.com/metal3d/idok/tunnel/go.crypto/ssh"
	"io/ioutil"
	"log"
)

// Parse local ssh private key to get signer
func parseSSHKeys(keyfile string) (ssh.Signer, error) {
	content, err := ioutil.ReadFile(keyfile)
	private, err := ssh.ParsePrivateKey(content)
	if err != nil {
		log.Println("Error: unable to parse private key")
	}
	return private, err
}
