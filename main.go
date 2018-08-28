package main

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
)

type Option struct {
	Name  string
	Value interface{}
}

func Ask(label string, opts []Option) (opt *Option, err error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "➜ {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: fmt.Sprintf("✔ %s: {{ .Name | red | cyan }}", label),
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     opts,
		Templates: templates,
		Size:      4,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	opt = &opts[i]
	return
}

func MustAsk(label string, opts []Option) *Option {
	ans, err := Ask(label, opts)
	if err != nil {
		log.Fatal(err)
	}
	return ans
}

func GenOpts(values []int, thing string) []Option {
	var opts []Option

	for _, value := range values {
		name := fmt.Sprintf("%d %s", value, thing)
		if value != 1 {
			name += "s"
		}

		opts = append(opts, Option{
			Name:  name,
			Value: value,
		})
	}

	return opts
}

func main() {
	ans := MustAsk("Number of resource chunks",
		GenOpts([]int{1, 2, 4, 8}, "chunk"))
	chunks := ans.Value.(int)

	ans = MustAsk("CPU cores per chunk",
		GenOpts([]int{1, 2, 4, 8, 16, 20, 24, 28}, "core"))
	cores := ans.Value.(int)

	fmt.Println(chunks, cores)
}
