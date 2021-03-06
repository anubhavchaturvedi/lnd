package lnwire

// ChannelID represent the set of data which is needed to retrieve all
// necessary data to validate the channel existence.
type ChannelID struct {
	// BlockHeight is the height of the block where funding transaction
	// located.
	//
	// NOTE: This field is limited to 3 bytes.
	BlockHeight uint32

	// TxIndex is a position of funding transaction within a block.
	//
	// NOTE: This field is limited to 3 bytes.
	TxIndex uint32

	// TxPosition indicating transaction output which pays to the
	// channel.
	TxPosition uint16
}

// NewChanIDFromInt returns a new ChannelID which is the decoded version of the
// compact channel ID encoded within the uint64. The format of the compact
// channel ID is as follows: 3 bytes for the block height, 3 bytes for the
// transaction index, and 2 bytes for the output index.
func NewChanIDFromInt(chanID uint64) ChannelID {
	return ChannelID{
		BlockHeight: uint32(chanID >> 40),
		TxIndex:     uint32(chanID>>16) & 0xFFFFFF,
		TxPosition:  uint16(chanID),
	}
}

// ToUint64 converts the ChannelID into a compact format encoded within a
// uint64 (8 bytes).
func (c *ChannelID) ToUint64() uint64 {
	// TODO(roasbeef): explicit error on overflow?
	return ((uint64(c.BlockHeight) << 40) | (uint64(c.TxIndex) << 16) |
		(uint64(c.TxPosition)))
}
