package infrastructure

import (
	"fmt"
	"strings"

	"github.com/zchase/jacuik/pkg/utils"
)

type ResourceOutput struct {
	Hash   string
	URN    string
	Name   string
	Status string
}

type InfrastructureOutput struct {
	WriteChannel chan ResourceOutput
}

func (i InfrastructureOutput) Write(msg []byte) (int, error) {
	msgParts := strings.Split(string(msg), " ")

	if len(msgParts) == 7 {

		urn := msgParts[3]
		name := msgParts[4]
		status := msgParts[5]

		// The Pulumi Stack resource never reports back as created from the streaming
		// output so lets ignore that row from our table.
		if urn == "pulumi:pulumi:Stack" {
			return len(msg), nil
		}

		var colorfulStatus string
		switch status {
		case "creating", "deleting", "updating", "update":
			colorfulStatus = utils.TextColor(status, "#f7bf2a")
			break
		case "created", "updated", "create":
			colorfulStatus = utils.TextColor(status, "#25a78b")
			break
		case "deleted", "delete":
			colorfulStatus = utils.TextColor(status, "#e53e3e")
		default:
			colorfulStatus = status
		}

		o := ResourceOutput{
			Hash:   utils.HashStringMD5(fmt.Sprintf("%s%s", status, urn)),
			URN:    urn,
			Name:   name,
			Status: colorfulStatus,
		}

		i.WriteChannel <- o
	}

	return len(msg), nil
}
