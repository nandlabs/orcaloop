package runtime

import "oss.nandlabs.io/golly/uuid"

func CreateId() string {
	uid, _ := uuid.V4()
	if uid != nil {

	}
	return uid.String()
}
