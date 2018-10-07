package chord

// Num + Key + Node ID + IP address == 100 bytes (REST) == 4 bytes + 16 bytes + 16 bytes + 4bytes (binary) = 40bytes
const findSuccessorMessageSizeREST = int64(100)

const findSuccessorMessageResponseSizeREST = int64(30)
