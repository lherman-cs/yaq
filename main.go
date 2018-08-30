package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

type Spec struct {
	Chunks   int
	Cores    int
	Mem      int
	GPUs     int
	GPUModel string
	Walltime string
}

type Option struct {
	Name  string
	Value interface{}
}

func Select(label string, opts []Option) (opt *Option, err error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "➜ {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: fmt.Sprintf("✔ %s: {{ .Name | yellow }}", label),
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

func MustSelect(label string, opts []Option) *Option {
	ans, err := Select(label, opts)
	if err != nil {
		log.Fatal(err)
	}
	return ans
}

func GenOpts(values []int, thing string, autoPlural bool) []Option {
	var opts []Option

	for _, value := range values {
		name := fmt.Sprintf("%d %s", value, thing)
		if autoPlural && value != 1 && value != 0 {
			name += "s"
		}

		opts = append(opts, Option{
			Name:  name,
			Value: value,
		})
	}

	return opts
}

func MustPromptWalltime() string {
	validate := func(input string) error {
		regex := regexp.MustCompile("^[0-9]{2}:[0-9]{2}:[0-9]{2}$")
		if !regex.MatchString(input) {
			return fmt.Errorf("Format: HH:mm:ss. e.g: 04:00:00 means 4 hours")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Walltime (HH:mm:ss)",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func Request(spec Spec) error {
	command := "qsub"
	specs := []string{
		fmt.Sprintf("select=%d", spec.Chunks),
		fmt.Sprintf("ncpus=%d", spec.Cores),
		fmt.Sprintf("mem=%dgb", spec.Mem),
		fmt.Sprintf("ngpus=%d", spec.GPUs),
	}
	if spec.GPUs > 0 {
		specs = append(specs, fmt.Sprintf("gpu_model=%s", spec.GPUModel))
	}

	hw := fmt.Sprintf("%s,walltime=%s", strings.Join(specs, ":"), spec.Walltime)
	args := []string{"-I", "-l", hw}

	fmt.Println("Generated:", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	ans := MustSelect("Number of resource chunks",
		GenOpts([]int{1, 2, 4, 8}, "chunk", true))
	chunks := ans.Value.(int)

	ans = MustSelect("CPU cores per chunk",
		GenOpts([]int{1, 2, 4, 8, 16, 20, 24, 28}, "core", true))
	cores := ans.Value.(int)

	ans = MustSelect("Amount of memory per chunk",
		GenOpts([]int{1, 2, 4, 6, 14, 30, 62, 120}, "GB", false))
	mem := ans.Value.(int)

	ans = MustSelect("Number of GPUs per chunk",
		GenOpts([]int{0, 1, 2}, "GPU", false))
	gpus := ans.Value.(int)

	var gpuModel string
	if gpus > 0 {
		ans = MustSelect("GPU Model", []Option{
			{Name: "m2075", Value: "m2075"},
			{Name: "m2070q", Value: "m2070q"},
			{Name: "k20", Value: "k20"},
			{Name: "m40", Value: "k40"},
			{Name: "p100", Value: "p100"},
		})
		gpuModel = ans.Value.(string)
	}

	walltime := MustPromptWalltime()

	spec := Spec{
		Chunks:   chunks,
		Cores:    cores,
		Mem:      mem,
		GPUs:     gpus,
		GPUModel: gpuModel,
		Walltime: walltime,
	}
	if err := Request(spec); err != nil {
		log.Fatal(err)
	}
}
