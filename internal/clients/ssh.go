package clients

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHKeyPrivate struct {
	priv *rsa.PrivateKey
}

func (k *SSHKeyPrivate) Generate() error {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	k.priv = priv
	return nil
}

func (k *SSHKeyPrivate) Load(str string) error {
	block, _ := pem.Decode([]byte(str))
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	k.priv = priv
	return nil
}

func (k *SSHKeyPrivate) LoadFile(file string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return k.Load(string(bytes))
}

func (k *SSHKeyPrivate) StorePublic() (string, error) {
	signer, err := ssh.NewSignerFromKey(k.priv)
	if err != nil {
		return "", err
	}
	return string(ssh.MarshalAuthorizedKey(signer.PublicKey())), nil
}

func (k *SSHKeyPrivate) Store() string {
	privDER := x509.MarshalPKCS1PrivateKey(k.priv)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privatePEM := pem.EncodeToMemory(&privBlock)
	return string(privatePEM)
}

func (k *SSHKeyPrivate) StoreFile(file string) error {
	str := k.Store()
	return os.WriteFile(file, []byte(str), 0o600)
}

func (k *SSHKeyPrivate) Execute(host string, port int, cmd string) (string, error) {
	signer, err := ssh.NewSignerFromKey(k.priv)
	if err != nil {
		return "", err
	}
	config := ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: time.Second * 10,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), &config)
	if err != nil {
		return "", err
	}
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("%w\n%s\n", err, string(output))
	}

	return string(output), nil
}
