package statediagram

import (
	"fmt"
	"strings"
)

type State struct {
	Start       bool
	End         bool
	Name        string
	Description string
	Next        *State
	Starts      []string
}

func (root *State) String() string {
	builder := strings.Builder{}

	name := strings.ReplaceAll(root.Name, "-", "_")
	//desc := strings.ReplaceAll(root.Description, "-", "_")

	starter := strings.Builder{}
	//starter.WriteString("stateDiagram-v2\n")

	if root.Start {
		if len(root.Starts) > 0 {
			for _, s := range root.Starts {
				sn := strings.ReplaceAll(s, "-", "_")
				starter.WriteString(fmt.Sprintf("\t%s --> %s : %s\n", sn, name, root.Description))
			}
		} else {
			starter.WriteString(fmt.Sprintf("\t[*] --> %s : %s\n", name, root.Description))
		}

		starter.WriteString(fmt.Sprintf("\tstate %s {\n", name))

	}

	idBuilder := strings.Builder{}

	stateIdPrefix := strings.ToLower(name)

	// go through all states until done
	c := root.Next

	if c == nil {
		return builder.String()
	}

	for {
		name := strings.ReplaceAll(c.Name, "-", "_")

		if c.Start {
			builder.WriteString("\t\t[*] --> ")
			builder.WriteString(stateIdPrefix + name)
			builder.WriteString("\n")
		}

		if c.End {
			builder.WriteString("\t\t")
			builder.WriteString(stateIdPrefix + name)
			builder.WriteString(" --> [*]")
			builder.WriteString("\n")
			builder.WriteString("\t}\n")
		}

		idBuilder.WriteString(fmt.Sprintf("\t\t%s%s : %s\n", stateIdPrefix, name, name))

		n := c.Next
		if n == nil {
			break
		}

		nn := strings.ReplaceAll(n.Name, "-", "_")

		builder.WriteString("\t\t")
		builder.WriteString(stateIdPrefix + name)
		builder.WriteString(" --> ")
		builder.WriteString(stateIdPrefix + nn)
		builder.WriteString("\n")

		c = n
	}

	starter.WriteString(idBuilder.String())
	starter.WriteString(builder.String())

	return starter.String()
}
