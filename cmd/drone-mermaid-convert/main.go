package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kanopy-platform/drone-mermaid-convert/pkg/statediagram"
	"gopkg.in/yaml.v3"
)

type Trigger struct {
	Branch []string `yaml:"branch"`
	Event  []string `yaml:"event"`
	Target []string `yaml:"target"`
}

type Pipeline struct {
	Name      string   `yaml:"name"`
	Kind      string   `yaml:"kind"`
	Trigger   *Trigger `yaml:"trigger"`
	Steps     []*Step  `yaml:"steps"`
	DependsOn []string `yaml:"depends_on"`
}

type Step struct {
	Name string `yaml:"name"`
}

type Drone struct {
	Pipelines []*Pipeline
}

func main() {
	if len(os.Args[1:]) != 1 {
		panic("Usage: drone-mermaid-convert <path to drone file>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	d := yaml.NewDecoder(f)

	drone := &Drone{
		Pipelines: []*Pipeline{},
	}

	for {
		// create new spec here
		p := &Pipeline{}
		// pass a reference to spec reference
		err := d.Decode(&p)
		// check it was parsed
		if p == nil {
			continue
		}
		// break the loop in case of EOF
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}

		drone.Pipelines = append(drone.Pipelines, p)
	}

	diagrames := []*statediagram.State{}

	for _, p := range drone.Pipelines {

		b := strings.Builder{}
		if len(p.Trigger.Branch) > 0 {
			b.WriteString(strings.Join(p.Trigger.Branch, " "))
		}

		if len(p.Trigger.Target) > 0 {
			b.WriteString(strings.Join(p.Trigger.Target, " "))
		}

		if (len(p.Trigger.Branch) > 0 || len(p.Trigger.Target) > 0) && len(p.Trigger.Event) > 0 {
			b.WriteString(" <- ")
		}

		if len(p.Trigger.Event) > 0 {
			b.WriteString(strings.Join(p.Trigger.Event, " "))
		}

		stateD := &statediagram.State{
			Start:       true,
			Name:        p.Name,
			Description: b.String(),
			Starts:      p.DependsOn,
		}

		var cp *statediagram.State
		cp = stateD
		for i, s := range p.Steps {
			ss := &statediagram.State{
				Name: s.Name,
			}

			if i == 0 {
				ss.Start = true
			}

			if i == len(p.Steps)-1 {
				ss.End = true
			}

			cp.Next = ss
			cp = ss
		}

		diagrames = append(diagrames, stateD)

	}
	fmt.Println("```mermaid")
	fmt.Println("stateDiagram-v2") // TODO needs to move out of here
	for _, d := range diagrames {
		fmt.Println(d.String())

	}
	fmt.Println("```")
}
