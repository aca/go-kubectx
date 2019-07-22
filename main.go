package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sahilm/fuzzy"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	err := kubens(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}

func kubens(query string) error {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeconfig, err := ioutil.ReadFile(rules.GetDefaultFilename())
	if err != nil {
		return err
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	var nsList []string
	for _, item := range namespaces.Items {
		nsList = append(nsList, item.ObjectMeta.Name)
	}

	matches := fuzzy.Find(query, nsList)

	// for _, match := range matches {
	// 	fmt.Println(highlightMatch(match))
	// }

	highlightMatches(matches)
	for i := len(matches) - 1; i >= 0; i-- {
		fmt.Println(i, matches[i].Str)
	}

	if matches.Len() == 0 {
		return fmt.Errorf("Failed to find matching namespace")
	}

	clientcmdConfig, err := rules.Load()
	if err != nil {
		return err
	}

	clientcmdConfig.Contexts[clientcmdConfig.CurrentContext].Namespace = matches[0].Str
	clientcmd.ModifyConfig(rules, *clientcmdConfig, false)
	return nil
}

// func highlightMatch(match fuzzy.Match) string {
// 	red := color.New(color.FgRed)
// 	midx := 0
// 	var sb strings.Builder
// 	for i, r := range match.Str {
// 		if midx < len(match.MatchedIndexes) && i == match.MatchedIndexes[midx] {
// 			sb.WriteString(red.Sprint(string(r)))
// 			midx++
// 		} else {
// 			sb.WriteRune(r)
// 		}
// 	}
// 	return sb.String()
// }

func highlightMatches(matches fuzzy.Matches) {
	red := color.New(color.FgRed)

	for k, match := range matches {
		var sb strings.Builder
		midx := 0
		for i, r := range match.Str {
			if midx < len(match.MatchedIndexes) && i == match.MatchedIndexes[midx] {
				sb.WriteString(red.Sprint(string(r)))
				midx++
			} else {
				sb.WriteRune(r)
			}
		}
		matches[k].Str = sb.String()
		fmt.Println(sb.String())
	}
	return
}
