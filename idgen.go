package hellorm

import (
	"log"

	"github.com/nu7hatch/gouuid"
)

func MkID() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	return id.String()
}
