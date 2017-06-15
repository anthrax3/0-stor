package reservation

import (
	"github.com/zero-os/0-stor/store/goraml"
	"gopkg.in/validator.v2"
)

type Reservation struct {
	AdminId      string          `json:"adminId" validate:"regexp=^\w+$,nonzero"`
	Created      goraml.DateTime `json:"created" validate:"nonzero"`
	ExpireAt     goraml.DateTime `json:"expireAt" validate:"nonzero"`
	Id           string          `json:"id" validate:"regexp=^\w+$,nonzero"`
	SizeReserved int64           `json:"sizeReserved" validate:"min=1,multipleOf=1,nonzero"`
	SizeUsed     int64           `json:"sizeUsed" validate:"min=1,nonzero"`
	Updated      goraml.DateTime `json:"updated" validate:"nonzero"`
}

func (s Reservation) Validate() error {

	return validator.Validate(s)
}
