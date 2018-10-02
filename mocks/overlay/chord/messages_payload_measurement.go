package chord

func findSuccessorMessageSize() int {
	// Num + Key + Node ID + IP address
	return 4 + 176 + 176 + 13
}
