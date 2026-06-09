package network

import "fmt"

type DiscReason uint

const (
	DiscRequested DiscReason = iota
	DiscNetworkError
	DiscAlreadyConnected
	DiscInvalidIdentity
	DiscQuitting
	DiscUnexpectedIdentity
	DiscSelf
	DiscReadTimeout
	DiscUnknownError
)

var discReasonToString = [...]string{
	DiscRequested:          "Disconnect requested",
	DiscNetworkError:       "Network error",
	DiscAlreadyConnected:   "Already connected",
	DiscInvalidIdentity:    "Invalid node identity",
	DiscQuitting:           "Client quitting",
	DiscUnexpectedIdentity: "Unexpected identity",
	DiscSelf:               "Connected to self",
	DiscReadTimeout:        "Read timeout",
	DiscUnknownError:       "Unkonwn error",
}

func (d DiscReason) String() string {
	if len(discReasonToString) < int(d) {
		return fmt.Sprintf("Unknown Reason(%d)", d)
	}
	return discReasonToString[d]
}

func (d DiscReason) Error() string {
	return d.String()
}

func discReasonForError(err error) DiscReason {
	if reason, ok := err.(DiscReason); ok {
		return reason
	} else {
		return DiscUnknownError
	}
}
