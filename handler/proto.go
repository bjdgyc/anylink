package handler

type Payload struct {
	ptype byte
	data  []byte
}

/*
   var header = []byte{'S', 'T', 'F', 0x01, 0, 0, 0x00, 0}
   https://tools.ietf.org/html/draft-mavrogiannopoulos-openconnect-02#section-2.2

   +---------------------+---------------------------------------------+
   |         byte        | value                                       |
   +---------------------+---------------------------------------------+
   |          0          | fixed to 0x53 (S)                           |
   |                     |                                             |
   |          1          | fixed to 0x54 (T)                           |
   |                     |                                             |
   |          2          | fixed to 0x46 (F)                           |
   |                     |                                             |
   |          3          | fixed to 0x01                               |
   |                     |                                             |
   |         4-5         | The length of the packet that follows this  |
   |                     | header in big endian order                  |
   |                     |                                             |
   |          6          | The type of the payload that follows (see   |
   |                     | Table 3 for available types)                |
   |                     |                                             |
   |          7          | fixed to 0x00                               |
   +---------------------+---------------------------------------------+


   The available payload types are listed in Table 3.
   +---------------------+---------------------------------------------+
   |        Value        | Description                                 |
   +---------------------+---------------------------------------------+
   |         0x00        | DATA: the TLS record packet contains an     |
   |                     | IPv4 or IPv6 packet                         |
   |                     |                                             |
   |         0x03        | DPD-REQ: used for dead peer detection. Once |
   |                     | sent the peer should reply with a DPD-RESP  |
   |                     | packet, that has the same contents as the   |
   |                     | original request.                           |
   |                     |                                             |
   |         0x04        | DPD-RESP: used as a response to a           |
   |                     | previously received DPD-REQ.                |
   |                     |                                             |
   |         0x05        | DISCONNECT: sent by the client (or server)  |
   |                     | to terminate the session.  No data is       |
   |                     | associated with this request. The session   |
   |                     | will be invalidated after such request.     |
   |                     |                                             |
   |         0x07        | KEEPALIVE: sent by any peer. No data is     |
   |                     | associated with this request.               |
   |                     |                                             |
   |         0x08        | COMPRESSED DATA: a Data packet which is     |
   |                     | compressed prior to encryption.             |
   |                     |                                             |
   |         0x09        | TERMINATE: sent by the server to indicate   |
   |                     | that the server is shutting down. No data   |
   |                     | is associated with this request.            |
   +---------------------+---------------------------------------------+
*/
