package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/aca/go-kubectx/config"
	"github.com/aca/go-kubectx/fzfutil"
	"github.com/spf13/pflag"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/client-go/tools/clientcmd"
)

var cfg *config.Config
var rules *clientcmd.ClientConfigLoadingRules
var clientcmdCfg *clientcmdapi.Config

func main() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg, err = config.ReadCfg()
	if err != nil {
		log.Fatal(err)
	}

	fs := pflag.NewFlagSet("kubectx", pflag.ContinueOnError)
	current := fs.BoolP("current", "c", false, "show the current context")
	help := fs.BoolP("help", "h", false, "help")
	fs.Parse(os.Args[1:])

	if *help {
		fmt.Printf(`USAGE:
  kubectx                : switch context(interactive)
  kubectx <QUERY>        : switch context which closest match with QUERY(if failed to match, interactive mode)
  kubectx -              : switch to previous context (~/.config/kubectx/config.json)
  kubectx -c, --current  : show current context name
  kubectx -h,--help      : help
`)
		return
	}

	rules = clientcmd.NewDefaultClientConfigLoadingRules()
	b, err := ioutil.ReadFile(rules.GetDefaultFilename())
	if err != nil {
		log.Fatal(err)
	}

	clientcmdCfg, err = clientcmd.Load(b)
	if err != nil {
		log.Fatal(err)
	}

	var query string

	if *current {
		fmt.Println(clientcmdCfg.CurrentContext)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "-" {
		query = cfg.LastContext
	} else if len(os.Args) > 1 {
		query = os.Args[1]
	}

	err = kubectx(query)
	if err != nil {
		fmt.Printf("Err: %s", err)
		os.Exit(1)
	}
}

func kubectx(query string) error {
	var err error

	_, ok := clientcmdCfg.Contexts[query]
	if ok {
		return modifyConfig(query)
	}

	listNS := func(in io.WriteCloser) {
		for key, _ := range clientcmdCfg.Contexts {
			fmt.Fprintln(in, key)
		}
	}

	fzfopt := []string{"--select-1", "--query", query}

	result, err := fzfutil.FZF(listNS, fzfopt...)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return errors.New("no context selected")
	}

	return modifyConfig(result[0])
}

func modifyConfig(to string) (err error) {
  curCtx := clientcmdCfg.CurrentContext
	defer func() {
		if err == nil {
			if curCtx != to {
				cfg.LastContext = curCtx
				config.WriteCfg(cfg)
			}
			fmt.Printf("Switched to context \"%s\"\n", to)
		}
	}()
	clientcmdCfg.CurrentContext = to
	return clientcmd.ModifyConfig(rules, *clientcmdCfg, false)
}
