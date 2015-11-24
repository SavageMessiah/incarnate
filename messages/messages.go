package messages

type Response struct {
	Body    string
	Context string
}

//Command is the world's dumbest tagged union
type Command struct {
	Complete struct {
		Request  *string
		Response *[]string
	}
	Command struct {
		Request  *string
		Response *Response
	}
	Broadcast *string
}
