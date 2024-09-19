# Zeppelin PacketLog
Super simple packet logger plugin for zeppelin

The command is the same as ProtocolLib's: 
`/packetlog <protocol> <sender> <id>`

Protocol: can be handshake|status|login|configuration|play

Sender: server|client

ID: Base-10 packet id

Packets in the log are raw, unencrypted, uncompressed with no length or id
