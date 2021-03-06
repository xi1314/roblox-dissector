package peer

import "github.com/gskartwii/roblox-dissector/datamodel"

// ID_CREATE_INSTANCE
type Packet83_02 struct {
	// The instance that was created
	Child *datamodel.Instance
}

func (thisBitstream *extendedReader) DecodePacket83_02(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	result, err := decodeReplicationInstance(reader, thisBitstream, layers)
	return &Packet83_02{result}, err
}

func (layer *Packet83_02) Serialize(writer PacketWriter, stream *extendedWriter) error {
	return serializeReplicationInstance(layer.Child, writer, stream)
}

func (Packet83_02) Type() uint8 {
	return 2
}
func (Packet83_02) TypeString() string {
	return "ID_REPLIC_NEW_INSTANCE"
}

func (layer *Packet83_02) String() string {
	return "ID_REPLIC_NEW_INSTANCE: " + layer.Child.GetFullName()
}
